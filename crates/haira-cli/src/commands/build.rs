//! Build command - compile a Haira file to a native binary.

use haira_ai::{AIConfig, AIEngine, AIError};
use haira_ast::{Item, ItemKind, SourceFile, Type};
use haira_cir::{
    CIRFunction, CIROperation, CIRType, CIRValue, CallSiteInfo, FieldDefinition,
    InterpretationContext, TypeDefinition,
};
use haira_codegen::{cir_to_function_def, compile_to_executable, CodegenOptions};
use haira_parser::parse;
use std::fs;
use std::path::Path;

pub fn run(
    file: &Path,
    output: Option<&Path>,
    interpret_ai: bool,
    use_ollama: bool,
    ollama_model: &str,
    mock_ai: bool,
) -> miette::Result<()> {
    let source =
        fs::read_to_string(file).map_err(|e| miette::miette!("Failed to read file: {}", e))?;

    eprintln!("Compiling: {}", file.display());

    let result = parse(&source);

    // Report parse errors
    if !result.errors.is_empty() {
        for err in &result.errors {
            eprintln!("Parse error: {}", err);
        }
        return Err(miette::miette!("{} parse error(s)", result.errors.len()));
    }

    // Check if there are AI blocks that need interpretation
    let ai_block_indices: Vec<usize> = result
        .ast
        .items
        .iter()
        .enumerate()
        .filter_map(|(i, item)| {
            if let ItemKind::AiFunctionDef(_) = &item.node {
                Some(i)
            } else {
                None
            }
        })
        .collect();

    let mut ast = result.ast;

    if !ai_block_indices.is_empty() {
        if mock_ai {
            // Use mock AI interpretation for testing
            eprintln!(
                "Found {} AI block(s) - using mock interpretation...",
                ai_block_indices.len()
            );

            for &idx in &ai_block_indices {
                let ai_block = match &ast.items[idx].node {
                    ItemKind::AiFunctionDef(block) => block.clone(),
                    _ => continue,
                };

                let name = ai_block
                    .name
                    .as_ref()
                    .map(|n| n.node.to_string())
                    .unwrap_or_else(|| format!("__ai_anon_{}", idx));

                eprintln!("  Generating mock for: {}", name);

                // Extract parameters
                let params: Vec<(String, String)> = ai_block
                    .params
                    .iter()
                    .map(|p| {
                        let ty =
                            p.ty.as_ref()
                                .map(|t| type_to_string(&t.node))
                                .unwrap_or_else(|| "any".to_string());
                        (p.name.node.to_string(), ty)
                    })
                    .collect();

                // Extract return type
                let return_type = ai_block.return_ty.as_ref().map(|t| type_to_string(&t.node));

                // Generate mock CIR
                let cir_func =
                    generate_mock_cir(&name, &params, return_type.as_deref(), &ai_block.intent);

                // Convert CIR to AST FunctionDef
                match cir_to_function_def(&cir_func) {
                    Ok(func_def) => {
                        let span = ast.items[idx].span;
                        ast.items[idx] = Item {
                            node: ItemKind::FunctionDef(func_def),
                            span,
                        };
                        eprintln!("    Generated: {}", name);
                    }
                    Err(e) => {
                        return Err(miette::miette!(
                            "Failed to convert mock CIR to AST for '{}': {}",
                            name,
                            e
                        ));
                    }
                }
            }

            eprintln!("All AI blocks processed with mock implementations.\n");
        } else if use_ollama {
            // Use local Ollama for AI interpretation
            eprintln!(
                "Found {} AI block(s) - using Ollama ({})...",
                ai_block_indices.len(),
                ollama_model
            );

            // Build interpretation context from the AST
            let context = build_interpretation_context(&ast, file);

            // Initialize AI engine with Ollama backend
            let config = AIConfig::default();
            let mut engine = AIEngine::with_ollama(config, Some(ollama_model));

            // Check Ollama availability
            let runtime = tokio::runtime::Runtime::new()
                .map_err(|e| miette::miette!("Failed to create async runtime: {}", e))?;

            runtime
                .block_on(async { engine.check_availability().await })
                .map_err(|e| {
                    miette::miette!(
                        "Ollama not available: {}\n\n\
                     Make sure Ollama is running:\n\
                       1. Install Ollama: https://ollama.ai\n\
                       2. Start the server: ollama serve\n\
                       3. Pull a model: ollama pull {}\n\n\
                     Or use --mock-ai for testing with stub implementations.",
                        e,
                        ollama_model
                    )
                })?;

            eprintln!("  Connected to Ollama server");

            // Process each AI block
            for &idx in &ai_block_indices {
                let ai_block = match &ast.items[idx].node {
                    ItemKind::AiFunctionDef(block) => block.clone(),
                    _ => continue,
                };

                let name = ai_block
                    .name
                    .as_ref()
                    .map(|n| n.node.to_string())
                    .unwrap_or_else(|| format!("__ai_anon_{}", idx));

                eprintln!("  Interpreting: {} ...", name);

                // Extract parameters as (name, type) pairs
                let params: Vec<(String, String)> = ai_block
                    .params
                    .iter()
                    .map(|p| {
                        let ty =
                            p.ty.as_ref()
                                .map(|t| type_to_string(&t.node))
                                .unwrap_or_else(|| "any".to_string());
                        (p.name.node.to_string(), ty)
                    })
                    .collect();

                // Extract return type
                let return_type = ai_block.return_ty.as_ref().map(|t| type_to_string(&t.node));

                // Call AI engine with Ollama
                let cir_result = runtime.block_on(engine.interpret_intent(
                    Some(&name),
                    ai_block.intent.as_str(),
                    &params,
                    return_type.as_deref(),
                    context.clone(),
                ));

                match cir_result {
                    Ok(cir_func) => {
                        eprintln!("    Generated CIR for: {}", cir_func.name);

                        // Convert CIR to AST FunctionDef
                        match cir_to_function_def(&cir_func) {
                            Ok(func_def) => {
                                let span = ast.items[idx].span;
                                ast.items[idx] = Item {
                                    node: ItemKind::FunctionDef(func_def),
                                    span,
                                };
                                eprintln!("    Converted to AST: {}", name);
                            }
                            Err(e) => {
                                return Err(miette::miette!(
                                    "Failed to convert CIR to AST for '{}': {}",
                                    name,
                                    e
                                ));
                            }
                        }
                    }
                    Err(e) => {
                        let error_msg = format_ai_error(&e);
                        return Err(miette::miette!(
                            "Failed to interpret AI block '{}': {}",
                            name,
                            error_msg
                        ));
                    }
                }
            }

            eprintln!("All AI blocks interpreted successfully.\n");
        } else if interpret_ai {
            eprintln!(
                "Found {} AI block(s) to interpret (Claude API)...",
                ai_block_indices.len()
            );

            // Build interpretation context from the AST
            let context = build_interpretation_context(&ast, file);

            // Initialize AI engine with Claude backend
            let config = AIConfig::from_env();
            if !config.is_valid() {
                return Err(miette::miette!(
                    "ANTHROPIC_API_KEY environment variable not set.\n\
                     Set it to your Anthropic API key to enable AI interpretation.\n\n\
                     Alternatively:\n\
                     - Use --ollama for local AI with Ollama\n\
                     - Use --mock-ai for testing with stub implementations"
                ));
            }

            let mut engine = AIEngine::new(config);

            // Process each AI block
            let runtime = tokio::runtime::Runtime::new()
                .map_err(|e| miette::miette!("Failed to create async runtime: {}", e))?;

            for &idx in &ai_block_indices {
                let ai_block = match &ast.items[idx].node {
                    ItemKind::AiFunctionDef(block) => block.clone(),
                    _ => continue,
                };

                let name = ai_block
                    .name
                    .as_ref()
                    .map(|n| n.node.to_string())
                    .unwrap_or_else(|| format!("__ai_anon_{}", idx));

                eprintln!("  Interpreting: {}", name);

                // Extract parameters as (name, type) pairs
                let params: Vec<(String, String)> = ai_block
                    .params
                    .iter()
                    .map(|p| {
                        let ty =
                            p.ty.as_ref()
                                .map(|t| type_to_string(&t.node))
                                .unwrap_or_else(|| "any".to_string());
                        (p.name.node.to_string(), ty)
                    })
                    .collect();

                // Extract return type
                let return_type = ai_block.return_ty.as_ref().map(|t| type_to_string(&t.node));

                // Call AI engine
                let cir_result = runtime.block_on(engine.interpret_intent(
                    Some(&name),
                    ai_block.intent.as_str(),
                    &params,
                    return_type.as_deref(),
                    context.clone(),
                ));

                match cir_result {
                    Ok(cir_func) => {
                        eprintln!("    Generated CIR for: {}", cir_func.name);

                        // Convert CIR to AST FunctionDef
                        match cir_to_function_def(&cir_func) {
                            Ok(func_def) => {
                                // Replace AI block with generated function
                                let span = ast.items[idx].span;
                                ast.items[idx] = Item {
                                    node: ItemKind::FunctionDef(func_def),
                                    span,
                                };
                                eprintln!("    Converted to AST: {}", name);
                            }
                            Err(e) => {
                                return Err(miette::miette!(
                                    "Failed to convert CIR to AST for '{}': {}",
                                    name,
                                    e
                                ));
                            }
                        }
                    }
                    Err(e) => {
                        let error_msg = format_ai_error(&e);
                        return Err(miette::miette!(
                            "Failed to interpret AI block '{}': {}",
                            name,
                            error_msg
                        ));
                    }
                }
            }

            eprintln!("All AI blocks interpreted successfully.\n");
        } else {
            return Err(miette::miette!(
                "Source contains {} AI block(s) which require interpretation.\n\n\
                 Options:\n\
                 - Use --ollama for local AI with Ollama (recommended)\n\
                 - Use --interpret-ai for Claude API (requires ANTHROPIC_API_KEY)\n\
                 - Use --mock-ai for testing with stub implementations",
                ai_block_indices.len()
            ));
        }
    }

    // Determine output binary name
    let output_file = output.map(|p| p.to_path_buf()).unwrap_or_else(|| {
        let stem = file.file_stem().unwrap_or_default();
        let output_dir = Path::new(".output");
        // Create .output directory if it doesn't exist
        if !output_dir.exists() {
            let _ = fs::create_dir_all(output_dir);
        }
        output_dir.join(stem)
    });

    // Compile to native binary
    let options = CodegenOptions::default();
    compile_to_executable(&ast, &output_file, options)
        .map_err(|e| miette::miette!("Compilation error: {}", e))?;

    eprintln!("Built: {}", output_file.display());

    Ok(())
}

/// Format AI error for display.
fn format_ai_error(e: &AIError) -> String {
    match e {
        AIError::MissingApiKey => "ANTHROPIC_API_KEY not set".to_string(),
        AIError::LowConfidence {
            confidence,
            minimum,
        } => {
            format!(
                "AI confidence too low: {:.1}% (minimum: {:.1}%)",
                confidence * 100.0,
                minimum * 100.0
            )
        }
        AIError::InterpretationFailed(msg) => {
            format!("Interpretation failed: {}", msg)
        }
        AIError::Ollama(ollama_err) => {
            format!("Ollama error: {}", ollama_err)
        }
        _ => e.to_string(),
    }
}

/// Generate a mock CIR function for testing.
fn generate_mock_cir(
    name: &str,
    params: &[(String, String)],
    return_type: Option<&str>,
    intent: &str,
) -> CIRFunction {
    let mut func =
        CIRFunction::new(name).with_description(format!("Mock implementation: {}", intent));

    // Add parameters
    for (param_name, param_type) in params {
        func = func.with_param(param_name, CIRType::simple(param_type));
    }

    // Set return type
    let ret_type = return_type.unwrap_or("none");
    func = func.returning(parse_type_string(ret_type));

    // Generate appropriate mock body based on return type
    // Note: The Haira compiler currently uses i64 for all values internally,
    // so we need to return appropriate values that work with i64.
    func = match ret_type {
        "int" => func
            .with_op(CIROperation::Literal {
                value: CIRValue::Int(0),
                result: "_result".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::Ref("_result".to_string()),
            }),
        "float" => func
            .with_op(CIROperation::Literal {
                value: CIRValue::Float(0.0),
                result: "_result".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::Ref("_result".to_string()),
            }),
        "string" => {
            // Return 0 as a placeholder - string handling in codegen returns pointer
            // but function signatures use i64. Real AI would need proper string handling.
            func.with_op(CIROperation::Literal {
                value: CIRValue::Int(0),
                result: "_result".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::Ref("_result".to_string()),
            })
        }
        "bool" => func
            .with_op(CIROperation::Literal {
                value: CIRValue::Int(1), // true as i64
                result: "_result".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::Ref("_result".to_string()),
            }),
        "none" | "" => func
            .with_op(CIROperation::Literal {
                value: CIRValue::Int(0),
                result: "_result".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::Ref("_result".to_string()),
            }),
        t if t.starts_with('[') => {
            // List type - return 0 as placeholder (empty list pointer)
            func.with_op(CIROperation::Literal {
                value: CIRValue::Int(0),
                result: "_result".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::Ref("_result".to_string()),
            })
        }
        _ => {
            // For other types (structs, etc.), return 0 as placeholder
            // Real AI implementation would construct proper objects
            func.with_op(CIROperation::Literal {
                value: CIRValue::Int(0),
                result: "_result".to_string(),
            })
            .with_op(CIROperation::Return {
                value: CIRValue::Ref("_result".to_string()),
            })
        }
    };

    func
}

/// Parse a type string into CIRType.
fn parse_type_string(s: &str) -> CIRType {
    if s.starts_with('[') && s.ends_with(']') {
        let inner = &s[1..s.len() - 1];
        CIRType::list(parse_type_string(inner))
    } else if s.starts_with("Option<") && s.ends_with('>') {
        let inner = &s[7..s.len() - 1];
        CIRType::option(parse_type_string(inner))
    } else {
        CIRType::simple(s)
    }
}

/// Build interpretation context from the parsed AST.
fn build_interpretation_context(ast: &SourceFile, file: &Path) -> InterpretationContext {
    let mut types_in_scope = Vec::new();

    // Extract all type definitions from the AST
    for item in &ast.items {
        if let ItemKind::TypeDef(type_def) = &item.node {
            let fields = type_def
                .fields
                .iter()
                .map(|f| {
                    let ty =
                        f.ty.as_ref()
                            .map(|t| type_to_string(&t.node))
                            .unwrap_or_else(|| "any".to_string());
                    FieldDefinition {
                        name: f.name.node.to_string(),
                        ty,
                        optional: false, // TODO: detect optionality
                        default: None,
                    }
                })
                .collect();

            types_in_scope.push(TypeDefinition {
                name: type_def.name.node.to_string(),
                fields,
            });
        }
    }

    InterpretationContext {
        types_in_scope,
        call_site: CallSiteInfo {
            file: file.display().to_string(),
            line: 1,
            arguments: vec![],
            expected_return: None,
        },
        project_schema: Default::default(),
    }
}

/// Convert an AST Type to a string representation.
fn type_to_string(ty: &Type) -> String {
    match ty {
        Type::Named(name) => name.to_string(),
        Type::List(inner) => format!("[{}]", type_to_string(&inner.node)),
        Type::Map { key, value } => {
            format!(
                "{{{}:{}}}",
                type_to_string(&key.node),
                type_to_string(&value.node)
            )
        }
        Type::Option(inner) => format!("Option<{}>", type_to_string(&inner.node)),
        Type::Function { params, ret } => {
            let params_str = params
                .iter()
                .map(|p| type_to_string(&p.node))
                .collect::<Vec<_>>()
                .join(", ");
            format!("({}) -> {}", params_str, type_to_string(&ret.node))
        }
        Type::Union(variants) => variants
            .iter()
            .map(|v| type_to_string(&v.node))
            .collect::<Vec<_>>()
            .join(" | "),
        Type::Generic { name, args } => {
            let args_str = args
                .iter()
                .map(|a| type_to_string(&a.node))
                .collect::<Vec<_>>()
                .join(", ");
            format!("{}<{}>", name, args_str)
        }
    }
}
