//! Build command - compile a Haira file to a native binary.

use haira_ai::hif::{
    cir_function_to_hif_intent, hif_intent_to_cir_function, parse_hif, write_hif, HIFFile,
};
use haira_ai::{AIConfig, AIEngine, AIError};
use haira_ast::{Item, ItemKind, SourceFile, Spanned, Type};
use haira_cir::{
    CIRFunction, CIROperation, CIRType, CIRValue, CallSiteInfo, FieldDefinition,
    InterpretationContext, TypeDefinition,
};
use haira_codegen::{cir_to_function_def, compile_to_executable, CodegenOptions};
use haira_parser::parse;
use std::fs;
use std::path::Path;

pub(crate) fn run(
    file: &Path,
    output: Option<&Path>,
    use_ollama: bool,
    ollama_model: &str,
    use_local_ai: bool,
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

    // Load HIF cache file if it exists
    let hif_path = file.with_extension("hif");
    let mut hif_file = load_hif_file(&hif_path);
    let mut hif_modified = false;

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

                // Compute hash for cache lookup
                let intent_hash = compute_intent_hash(&name, &ai_block.intent);

                // Check HIF cache first
                if let Some(cached_intent) = hif_file.get_intent(&name) {
                    if cached_intent.hash == intent_hash {
                        eprintln!("  Using cached: {} (from .hif)", name);
                        let cir_func = hif_intent_to_cir_function(cached_intent);

                        match cir_to_function_def(&cir_func) {
                            Ok(func_def) => {
                                let span = ast.items[idx].span;
                                ast.items[idx] = Item {
                                    node: ItemKind::FunctionDef(func_def),
                                    span,
                                };
                                continue;
                            }
                            Err(e) => {
                                eprintln!("    Cache invalid, re-interpreting: {}", e);
                            }
                        }
                    } else {
                        eprintln!("  Cache stale for: {} (intent changed)", name);
                    }
                }

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

                        // Save to HIF cache
                        let hif_intent = cir_function_to_hif_intent(&cir_func, &intent_hash);
                        hif_file.add_intent(hif_intent);
                        hif_modified = true;

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

            // Save HIF file if modified
            if hif_modified {
                save_hif_file(&hif_path, &hif_file);
            }

            eprintln!("All AI blocks interpreted successfully.\n");
        } else if use_local_ai {
            // Use local llama.cpp for AI interpretation
            eprintln!(
                "Found {} AI block(s) - using local AI (llama.cpp)...",
                ai_block_indices.len()
            );

            // Build interpretation context from the AST
            let context = build_interpretation_context(&ast, file);

            // Initialize AI engine with local AI backend
            let config = AIConfig::default();
            let mut engine = AIEngine::with_local_ai(config, None);

            // Check local AI availability
            let runtime = tokio::runtime::Runtime::new()
                .map_err(|e| miette::miette!("Failed to create async runtime: {}", e))?;

            runtime
                .block_on(async { engine.check_availability().await })
                .map_err(|e| {
                    miette::miette!(
                        "Local AI not available: {}\n\n\
                     Make sure you have:\n\
                       1. Installed the model: haira model pull\n\
                       2. The llama-server binary in ~/.haira/bin/\n\n\
                     Or use --mock-ai for testing with stub implementations.",
                        e
                    )
                })?;

            // Start the local AI server
            eprintln!("  Starting local AI server...");
            runtime
                .block_on(async { engine.start_local_server().await })
                .map_err(|e| miette::miette!("Failed to start local AI server: {}", e))?;

            eprintln!("  Local AI server ready");

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

                // Compute hash for cache lookup
                let intent_hash = compute_intent_hash(&name, &ai_block.intent);

                // Check HIF cache first
                if let Some(cached_intent) = hif_file.get_intent(&name) {
                    if cached_intent.hash == intent_hash {
                        eprintln!("  Using cached: {} (from .hif)", name);
                        let cir_func = hif_intent_to_cir_function(cached_intent);

                        match cir_to_function_def(&cir_func) {
                            Ok(func_def) => {
                                let span = ast.items[idx].span;
                                ast.items[idx] = Item {
                                    node: ItemKind::FunctionDef(func_def),
                                    span,
                                };
                                continue;
                            }
                            Err(e) => {
                                eprintln!("    Cache invalid, re-interpreting: {}", e);
                            }
                        }
                    } else {
                        eprintln!("  Cache stale for: {} (intent changed)", name);
                    }
                }

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

                // Call AI engine with local AI
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

                        // Save to HIF cache
                        let hif_intent = cir_function_to_hif_intent(&cir_func, &intent_hash);
                        hif_file.add_intent(hif_intent);
                        hif_modified = true;

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
                                // Stop the server before returning error
                                let _ = engine.stop_local_server();
                                return Err(miette::miette!(
                                    "Failed to convert CIR to AST for '{}': {}",
                                    name,
                                    e
                                ));
                            }
                        }
                    }
                    Err(e) => {
                        // Stop the server before returning error
                        let _ = engine.stop_local_server();
                        let error_msg = format_ai_error(&e);
                        return Err(miette::miette!(
                            "Failed to interpret AI block '{}': {}",
                            name,
                            error_msg
                        ));
                    }
                }
            }

            // Stop the local AI server
            let _ = engine.stop_local_server();

            // Save HIF file if modified
            if hif_modified {
                save_hif_file(&hif_path, &hif_file);
            }

            eprintln!("All AI blocks interpreted successfully.\n");
        } else {
            return Err(miette::miette!(
                "Source contains {} AI block(s) which require interpretation.\n\n\
                 Options:\n\
                 - Use --local-ai for local AI with llama.cpp (recommended)\n\
                 - Use --ollama for local AI with Ollama\n\
                 - Use --mock-ai for testing with stub implementations",
                ai_block_indices.len()
            ));
        }
    }

    // Infer types for struct fields that don't have explicit type annotations
    // This uses AI to determine types based on field names
    let ast = infer_struct_field_types(ast, use_ollama, ollama_model, use_local_ai)?;

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
        AIError::LocalAI(local_err) => {
            format!("Local AI error: {}", local_err)
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

/// Infer types for struct fields that don't have explicit type annotations.
///
/// This function scans all struct definitions in the AST and uses AI to infer
/// types for fields that don't have type annotations.
fn infer_struct_field_types(
    mut ast: SourceFile,
    use_ollama: bool,
    ollama_model: &str,
    use_local_ai: bool,
) -> miette::Result<SourceFile> {
    // Find all structs with untyped fields
    let mut structs_needing_inference: Vec<(usize, String, Vec<String>)> = Vec::new();

    for (idx, item) in ast.items.iter().enumerate() {
        if let ItemKind::TypeDef(type_def) = &item.node {
            let untyped_fields: Vec<String> = type_def
                .fields
                .iter()
                .filter(|f| f.ty.is_none())
                .map(|f| f.name.node.to_string())
                .collect();

            if !untyped_fields.is_empty() {
                structs_needing_inference.push((
                    idx,
                    type_def.name.node.to_string(),
                    untyped_fields,
                ));
            }
        }
    }

    if structs_needing_inference.is_empty() {
        return Ok(ast);
    }

    eprintln!(
        "Found {} struct(s) with untyped fields - inferring types with AI...",
        structs_needing_inference.len()
    );

    // Initialize AI engine based on flags
    let config = AIConfig::default();
    let engine = if use_ollama {
        AIEngine::with_ollama(config, Some(ollama_model))
    } else if use_local_ai {
        AIEngine::with_local_ai(config, None)
    } else {
        eprintln!("  No AI backend specified, using default type inference");
        return Ok(apply_default_types(ast, &structs_needing_inference));
    };

    // Create async runtime
    let runtime = tokio::runtime::Runtime::new()
        .map_err(|e| miette::miette!("Failed to create async runtime: {}", e))?;

    // Infer types for each struct
    for (idx, struct_name, field_names) in &structs_needing_inference {
        eprintln!("  Inferring types for struct '{}'...", struct_name);

        let inferred_types = runtime.block_on(async {
            engine
                .infer_struct_field_types(struct_name, field_names)
                .await
        });

        match inferred_types {
            Ok(types) => {
                // Update the AST with inferred types
                if let ItemKind::TypeDef(ref mut type_def) = ast.items[*idx].node {
                    for field in &mut type_def.fields {
                        if field.ty.is_none() {
                            let field_name = field.name.node.to_string();
                            if let Some(type_str) = types.get(&field_name) {
                                let inferred_type = string_to_type(type_str);
                                field.ty = Some(Spanned {
                                    node: inferred_type,
                                    span: field.name.span,
                                });
                                eprintln!("    {} -> {}", field_name, type_str);
                            }
                        }
                    }
                }
            }
            Err(e) => {
                eprintln!(
                    "  Warning: Failed to infer types for '{}': {}",
                    struct_name, e
                );
                eprintln!("  Using default types (int for unknown fields)");
                // Apply default types for this struct
                if let ItemKind::TypeDef(ref mut type_def) = ast.items[*idx].node {
                    for field in &mut type_def.fields {
                        if field.ty.is_none() {
                            let default_type = infer_type_from_name(&field.name.node);
                            field.ty = Some(Spanned {
                                node: string_to_type(&default_type),
                                span: field.name.span,
                            });
                            eprintln!("    {} -> {} (default)", field.name.node, default_type);
                        }
                    }
                }
            }
        }
    }

    eprintln!("Type inference complete.\n");
    Ok(ast)
}

/// Apply default types to structs without using AI.
fn apply_default_types(
    mut ast: SourceFile,
    structs: &[(usize, String, Vec<String>)],
) -> SourceFile {
    for (idx, struct_name, _) in structs {
        eprintln!("  Applying default types for struct '{}'...", struct_name);
        if let ItemKind::TypeDef(ref mut type_def) = ast.items[*idx].node {
            for field in &mut type_def.fields {
                if field.ty.is_none() {
                    let default_type = infer_type_from_name(&field.name.node);
                    field.ty = Some(Spanned {
                        node: string_to_type(&default_type),
                        span: field.name.span,
                    });
                    eprintln!("    {} -> {}", field.name.node, default_type);
                }
            }
        }
    }
    ast
}

/// Infer a default type from a field name using simple heuristics.
fn infer_type_from_name(name: &str) -> String {
    let lower = name.to_lowercase();

    // Boolean patterns
    if lower.starts_with("is_")
        || lower.starts_with("has_")
        || lower.starts_with("can_")
        || lower.starts_with("should_")
        || lower == "active"
        || lower == "enabled"
        || lower == "visible"
        || lower == "valid"
        || lower == "done"
        || lower == "completed"
    {
        return "bool".to_string();
    }

    // Integer patterns
    if lower == "id"
        || lower.ends_with("_id")
        || lower == "age"
        || lower == "count"
        || lower == "quantity"
        || lower == "index"
        || lower == "size"
        || lower == "length"
        || lower == "year"
        || lower == "month"
        || lower == "day"
        || lower == "hour"
        || lower == "minute"
        || lower == "second"
    {
        return "int".to_string();
    }

    // Float patterns
    if lower == "price"
        || lower == "amount"
        || lower == "rate"
        || lower == "ratio"
        || lower == "percentage"
        || lower == "score"
        || lower == "weight"
        || lower == "height"
        || lower == "width"
        || lower == "latitude"
        || lower == "longitude"
        || lower.ends_with("_rate")
        || lower.ends_with("_ratio")
    {
        return "float".to_string();
    }

    // Default to string for names, descriptions, etc.
    "string".to_string()
}

/// Convert a type string to an AST Type.
fn string_to_type(s: &str) -> Type {
    match s {
        "int" | "i64" | "i32" => Type::Named(s.into()),
        "float" | "f64" | "f32" => Type::Named("float".into()),
        "bool" | "boolean" => Type::Named("bool".into()),
        "string" | "str" => Type::Named("string".into()),
        _ => Type::Named(s.into()),
    }
}

/// Load a HIF cache file if it exists.
fn load_hif_file(path: &Path) -> HIFFile {
    if path.exists() {
        match fs::read_to_string(path) {
            Ok(content) => match parse_hif(&content) {
                Ok(hif) => {
                    eprintln!("Loaded HIF cache: {}", path.display());
                    return hif;
                }
                Err(e) => {
                    eprintln!("Warning: Failed to parse HIF cache: {}", e);
                }
            },
            Err(e) => {
                eprintln!("Warning: Failed to read HIF cache: {}", e);
            }
        }
    }
    HIFFile::new()
}

/// Save a HIF cache file.
fn save_hif_file(path: &Path, hif: &HIFFile) {
    let content = write_hif(hif);
    match fs::write(path, &content) {
        Ok(_) => {
            eprintln!("Saved HIF cache: {}", path.display());
        }
        Err(e) => {
            eprintln!("Warning: Failed to save HIF cache: {}", e);
        }
    }
}

/// Compute a hash for an intent based on name and content.
fn compute_intent_hash(name: &str, intent: &str) -> String {
    use std::collections::hash_map::DefaultHasher;
    use std::hash::{Hash, Hasher};

    let mut hasher = DefaultHasher::new();
    name.hash(&mut hasher);
    intent.hash(&mut hasher);
    format!("{:x}", hasher.finish())
}
