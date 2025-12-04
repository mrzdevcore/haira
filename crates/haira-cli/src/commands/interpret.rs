//! Interpret command - test AI interpretation of function names.

use haira_ai::{AIConfig, AIEngine, InterpretationContext, TypeDefinition};
use haira_cir::{CallSiteInfo, FieldDefinition};
use std::path::Path;

pub(crate) async fn run(name: &str, context_file: Option<&Path>) -> miette::Result<()> {
    println!("Interpreting function: {}\n", name);

    // Load context if provided
    let context = if let Some(path) = context_file {
        load_context(path)?
    } else {
        default_context()
    };

    // Create AI engine
    let config = AIConfig::default();
    let mut engine = AIEngine::new(config);

    // Check if we can use pattern matching (no AI needed)
    println!("Checking pattern matching...");
    if engine.matches_pattern(name) {
        println!("  Matched by pattern! May not need AI call.\n");
    } else {
        println!("  No pattern match, would use AI interpretation.\n");
    }

    // Try AI interpretation
    println!("Calling AI for interpretation...\n");

    match engine.interpret(name, context).await {
        Ok(func) => {
            println!("Interpretation successful!\n");
            println!("Generated CIR:");
            println!("{}", serde_json::to_string_pretty(&func).unwrap());
        }
        Err(e) => {
            println!("AI interpretation failed: {}", e);
            println!("\nFalling back to pattern analysis...");
            analyze_name_patterns(name);
        }
    }

    Ok(())
}

fn load_context(path: &Path) -> miette::Result<InterpretationContext> {
    let content = std::fs::read_to_string(path)
        .map_err(|e| miette::miette!("Failed to read context file: {}", e))?;

    let context: InterpretationContext = serde_json::from_str(&content)
        .map_err(|e| miette::miette!("Failed to parse context JSON: {}", e))?;

    Ok(context)
}

fn default_context() -> InterpretationContext {
    // Create a default context with common types
    InterpretationContext {
        types_in_scope: vec![
            TypeDefinition {
                name: "User".to_string(),
                fields: vec![
                    FieldDefinition {
                        name: "id".to_string(),
                        ty: "Int".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "name".to_string(),
                        ty: "String".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "email".to_string(),
                        ty: "String".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "active".to_string(),
                        ty: "Bool".to_string(),
                        optional: false,
                        default: None,
                    },
                ],
            },
            TypeDefinition {
                name: "Post".to_string(),
                fields: vec![
                    FieldDefinition {
                        name: "id".to_string(),
                        ty: "Int".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "title".to_string(),
                        ty: "String".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "content".to_string(),
                        ty: "String".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "author_id".to_string(),
                        ty: "Int".to_string(),
                        optional: false,
                        default: None,
                    },
                ],
            },
        ],
        call_site: CallSiteInfo {
            file: "main.haira".to_string(),
            line: 1,
            arguments: vec![],
            expected_return: None,
        },
        project_schema: Default::default(),
    }
}

fn analyze_name_patterns(name: &str) {
    println!("Pattern Analysis:");
    println!("  Function name: {}", name);

    // Split by underscore
    let parts: Vec<&str> = name.split('_').collect();
    println!("  Parts: {:?}", parts);

    // Identify verb
    if let Some(verb) = parts.first() {
        let verb_analysis = match *verb {
            "get" => "Retrieval operation - fetches data",
            "set" => "Mutation operation - updates data",
            "create" | "make" | "new" => "Creation operation - creates new instance",
            "delete" | "remove" => "Deletion operation - removes data",
            "find" | "search" => "Search operation - locates data",
            "filter" => "Filter operation - selects subset",
            "map" | "transform" => "Transform operation - converts data",
            "validate" | "check" => "Validation operation - verifies data",
            "save" | "store" => "Persistence operation - stores data",
            "load" | "fetch" => "Loading operation - retrieves data",
            "send" | "emit" => "Emission operation - sends data/events",
            "parse" => "Parsing operation - interprets text",
            "format" => "Formatting operation - creates string",
            "sort" | "order" => "Sorting operation - orders data",
            "count" => "Counting operation - returns count",
            "sum" | "total" => "Aggregation operation - sums values",
            "is" | "has" | "can" => "Predicate - returns boolean",
            _ => "Unknown verb pattern",
        };
        println!("  Verb '{}': {}", verb, verb_analysis);
    }

    // Identify noun (likely type reference)
    if parts.len() > 1 {
        let nouns: Vec<&str> = parts[1..].to_vec();
        println!("  Nouns: {:?}", nouns);

        // Check for pluralization
        if let Some(last) = nouns.last() {
            if last.ends_with('s') && !last.ends_with("ss") {
                println!(
                    "  Note: '{}' appears plural - likely returns collection",
                    last
                );
            }
        }

        // Check for common modifiers
        for noun in &nouns {
            match *noun {
                "by" => println!("  'by' - indicates filtering/lookup criteria follows"),
                "with" => println!("  'with' - indicates inclusion of related data"),
                "all" => println!("  'all' - indicates full collection retrieval"),
                "active" | "inactive" => println!("  '{}' - status filter", noun),
                "recent" | "latest" => println!("  '{}' - temporal filter", noun),
                _ => {}
            }
        }
    }

    println!();
    println!("Suggested implementation approach:");
    suggest_implementation(&parts);
}

fn suggest_implementation(parts: &[&str]) {
    if parts.is_empty() {
        println!("  Unable to determine - name too short");
        return;
    }

    match parts[0] {
        "get" if parts.len() > 1 => {
            let target = parts[1..].join("_");
            if target.ends_with('s') {
                println!(
                    "  1. Query database for {} table",
                    &target[..target.len() - 1]
                );
                println!("  2. Apply any filters (by_*, active, etc.)");
                println!("  3. Return collection");
            } else {
                println!("  1. Query database for single {}", target);
                println!("  2. Return Option<{}>", capitalize(&target));
            }
        }
        "create" | "new" if parts.len() > 1 => {
            let target = parts[1..].join("_");
            println!("  1. Validate input parameters");
            println!("  2. Create new {} instance", target);
            println!("  3. Persist to storage if applicable");
            println!("  4. Return created instance");
        }
        "filter" if parts.len() > 1 => {
            println!("  1. Take collection as input");
            println!("  2. Apply predicate based on '{}'", parts[1..].join("_"));
            println!("  3. Return filtered collection");
        }
        "validate" | "check" if parts.len() > 1 => {
            let target = parts[1..].join("_");
            println!("  1. Check {} against validation rules", target);
            println!("  2. Return (Bool, ?Error)");
        }
        _ => {
            println!("  Analysis inconclusive - would require AI interpretation");
        }
    }
}

fn capitalize(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        None => String::new(),
        Some(first) => first.to_uppercase().chain(chars).collect(),
    }
}
