package haira

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// ---------------------------------------------------------------------------
// Embeddings
// ---------------------------------------------------------------------------

// VectorEmbed generates an embedding for a single text using the given provider.
// Panics on error (caught by Haira's try/catch via recover).
func VectorEmbed(provider *Provider, text string) []float32 {
	client := CreateOpenAIClient(provider)

	resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(provider.Model),
	})
	if err != nil {
		panic(fmt.Sprintf("embedding error: %v", err))
	}
	if len(resp.Data) == 0 {
		panic("embedding returned no data")
	}
	return resp.Data[0].Embedding
}

// VectorEmbedBatch generates embeddings for multiple texts.
// Panics on error (caught by Haira's try/catch via recover).
func VectorEmbedBatch(provider *Provider, texts []any) [][]float32 {
	strs := make([]string, len(texts))
	for i, t := range texts {
		strs[i] = fmt.Sprintf("%v", t)
	}

	client := CreateOpenAIClient(provider)

	resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: strs,
		Model: openai.EmbeddingModel(provider.Model),
	})
	if err != nil {
		panic(fmt.Sprintf("batch embedding error: %v", err))
	}

	results := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		results[i] = d.Embedding
	}
	return results
}

// ---------------------------------------------------------------------------
// pgvector Collection
// ---------------------------------------------------------------------------

// VectorCollection represents a pgvector-backed document collection.
type VectorCollection struct {
	DB         *DB
	Table      string
	Dimensions int
}

// VectorNewCollection creates or ensures a pgvector collection table exists.
// Creates: id SERIAL PRIMARY KEY, content TEXT, embedding vector(N), metadata JSONB, created_at TIMESTAMPTZ.
func VectorNewCollection(db *DB, name string, dimensions int) *VectorCollection {
	// Enable pgvector extension
	_, err := db.conn.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		panic(fmt.Sprintf("pgvector extension: %v", err))
	}

	// Create table if not exists
	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			embedding vector(%d),
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`, name, dimensions)
	_, err = db.conn.Exec(createSQL)
	if err != nil {
		panic(fmt.Sprintf("create collection %q: %v", name, err))
	}

	// Create index for cosine similarity search
	indexSQL := fmt.Sprintf(
		"CREATE INDEX IF NOT EXISTS %s_embedding_idx ON %s USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
		name, name,
	)
	// Ignore index creation errors (may fail if not enough rows yet)
	db.conn.Exec(indexSQL)

	return &VectorCollection{
		DB:         db,
		Table:      name,
		Dimensions: dimensions,
	}
}

// VectorInsert inserts a document with embedding and optional metadata.
// params: { content: string, embedding: []float32, metadata: map[string]any }
func VectorInsert(coll *VectorCollection, params map[string]any) {
	content, _ := params["content"].(string)

	// Convert embedding to pgvector format: [0.1,0.2,0.3]
	embeddingStr := formatVector(params["embedding"])

	// Convert metadata to JSON
	metadataJSON := "{}"
	if meta, ok := params["metadata"]; ok {
		if b, err := json.Marshal(meta); err == nil {
			metadataJSON = string(b)
		}
	}

	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (content, embedding, metadata) VALUES ($1, $2::vector, $3::jsonb)",
		coll.Table,
	)
	_, err := coll.DB.conn.Exec(insertSQL, content, embeddingStr, metadataJSON)
	if err != nil {
		panic(fmt.Sprintf("vector insert: %v", err))
	}
}

// VectorSearch performs cosine similarity search on the collection.
// params: { query: []float32, limit: int, filter: string (optional SQL WHERE clause) }
// Returns: [{ content: string, metadata: map, distance: float64 }]
func VectorSearch(coll *VectorCollection, params map[string]any) []map[string]any {
	queryStr := formatVector(params["query"])

	limit := 5
	if l, ok := params["limit"]; ok {
		switch v := l.(type) {
		case int:
			limit = v
		case float64:
			limit = int(v)
		}
	}

	// Build query with optional filter
	filter := ""
	if f, ok := params["filter"].(string); ok && f != "" {
		filter = " WHERE " + f
	}

	searchSQL := fmt.Sprintf(
		"SELECT content, metadata, embedding <=> $1::vector AS distance FROM %s%s ORDER BY distance LIMIT $2",
		coll.Table, filter,
	)

	rows, err := coll.DB.conn.Query(searchSQL, queryStr, limit)
	if err != nil {
		panic(fmt.Sprintf("vector search: %v", err))
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var content string
		var metadataRaw []byte
		var distance float64

		if err := rows.Scan(&content, &metadataRaw, &distance); err != nil {
			panic(fmt.Sprintf("vector search scan: %v", err))
		}

		var metadata map[string]any
		json.Unmarshal(metadataRaw, &metadata)
		if metadata == nil {
			metadata = map[string]any{}
		}

		results = append(results, map[string]any{
			"content":  content,
			"metadata": metadata,
			"distance": distance,
		})
	}

	return results
}

// VectorFormat formats search results as text suitable for LLM context.
func VectorFormat(results []map[string]any) string {
	if len(results) == 0 {
		return "No results found."
	}

	var sb strings.Builder
	for i, r := range results {
		content, _ := r["content"].(string)
		distance, _ := r["distance"].(float64)
		fmt.Fprintf(&sb, "[%d] (score: %.3f) %s\n", i+1, 1-distance, content)
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// formatVector converts an embedding ([]float32 or []any) to pgvector string format: [0.1,0.2,0.3]
func formatVector(v any) string {
	switch vec := v.(type) {
	case []float32:
		parts := make([]string, len(vec))
		for i, f := range vec {
			parts[i] = fmt.Sprintf("%g", f)
		}
		return "[" + strings.Join(parts, ",") + "]"
	case []any:
		parts := make([]string, len(vec))
		for i, f := range vec {
			parts[i] = fmt.Sprintf("%v", f)
		}
		return "[" + strings.Join(parts, ",") + "]"
	case []float64:
		parts := make([]string, len(vec))
		for i, f := range vec {
			parts[i] = fmt.Sprintf("%g", f)
		}
		return "[" + strings.Join(parts, ",") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
}
