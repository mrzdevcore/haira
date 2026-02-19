package haira

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"

	chromem "github.com/philippgille/chromem-go"
)

// LocalVectorCollection is a chromem-go backed vector collection for local development.
// It provides the same operations as pgvector but requires no external database.
type LocalVectorCollection struct {
	coll *chromem.Collection
	db   *chromem.DB
}

var localVectorDB *chromem.DB
var localDocCounter atomic.Int64

// getLocalVectorDB returns the shared in-memory chromem-go database, creating it if needed.
func getLocalVectorDB() *chromem.DB {
	if localVectorDB == nil {
		localVectorDB = chromem.NewDB()
	}
	return localVectorDB
}

// VectorNewLocalCollection creates or gets a chromem-go collection for local vector search.
func VectorNewLocalCollection(name string, dimensions int) *LocalVectorCollection {
	db := getLocalVectorDB()

	// Use GetOrCreateCollection with nil embedding func (we provide pre-computed embeddings)
	coll, err := db.GetOrCreateCollection(name, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("local vector collection %q: %v", name, err))
	}

	return &LocalVectorCollection{
		coll: coll,
		db:   db,
	}
}

// LocalVectorInsert inserts a document with a pre-computed embedding into the local collection.
// params: { content: string, embedding: []float32, metadata: map[string]any }
func LocalVectorInsert(coll *LocalVectorCollection, params map[string]any) {
	content, _ := params["content"].(string)

	embedding := toFloat32Slice(params["embedding"])
	if len(embedding) == 0 {
		panic("local vector insert: embedding is required")
	}

	// Convert metadata to map[string]string (chromem-go requirement)
	metaStr := make(map[string]string)
	if meta, ok := params["metadata"]; ok {
		switch m := meta.(type) {
		case map[string]any:
			for k, v := range m {
				metaStr[k] = fmt.Sprintf("%v", v)
			}
		case map[string]string:
			metaStr = m
		}
	}

	// Generate a unique document ID
	docID := fmt.Sprintf("doc_%d", localDocCounter.Add(1))

	doc := chromem.Document{
		ID:        docID,
		Content:   content,
		Embedding: embedding,
		Metadata:  metaStr,
	}

	err := coll.coll.AddDocument(context.Background(), doc)
	if err != nil {
		panic(fmt.Sprintf("local vector insert: %v", err))
	}
}

// LocalVectorSearch performs cosine similarity search on the local collection.
// params: { query: []float32, limit: int }
// Returns: [{ content: string, metadata: map, distance: float64 }]
func LocalVectorSearch(coll *LocalVectorCollection, params map[string]any) []map[string]any {
	queryEmbedding := toFloat32Slice(params["query"])
	if len(queryEmbedding) == 0 {
		panic("local vector search: query embedding is required")
	}

	limit := 5
	if l, ok := params["limit"]; ok {
		switch v := l.(type) {
		case int:
			limit = v
		case float64:
			limit = int(v)
		}
	}

	// chromem-go requires nResults > 0 and <= collection count
	count := coll.coll.Count()
	if count == 0 {
		return nil
	}
	if limit > count {
		limit = count
	}

	results, err := coll.coll.QueryEmbedding(context.Background(), queryEmbedding, limit, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("local vector search: %v", err))
	}

	var out []map[string]any
	for _, r := range results {
		// Convert chromem-go similarity (higher=better, range [-1,1]) to distance (lower=better)
		distance := 1.0 - float64(r.Similarity)

		// Convert metadata back to map[string]any
		meta := make(map[string]any, len(r.Metadata))
		for k, v := range r.Metadata {
			meta[k] = v
		}

		out = append(out, map[string]any{
			"content":  r.Content,
			"metadata": meta,
			"distance": distance,
		})
	}

	return out
}

// toFloat32Slice converts various numeric slice types to []float32.
func toFloat32Slice(v any) []float32 {
	switch vec := v.(type) {
	case []float32:
		return vec
	case []float64:
		out := make([]float32, len(vec))
		for i, f := range vec {
			out[i] = float32(f)
		}
		return out
	case []any:
		out := make([]float32, len(vec))
		for i, f := range vec {
			switch n := f.(type) {
			case float64:
				out[i] = float32(n)
			case float32:
				out[i] = n
			case int:
				out[i] = float32(n)
			case string:
				if parsed, err := strconv.ParseFloat(n, 32); err == nil {
					out[i] = float32(parsed)
				}
			}
		}
		return out
	default:
		return nil
	}
}
