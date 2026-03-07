package codegen

import (
	"fmt"
	"strings"

	"github.com/haira-lang/haira/internal/ast"
)

// ResolveStdlibCall checks if a call is a known stdlib function and returns Go code.
func ResolveStdlibCall(call ast.CallExpr) (string, bool) {
	// Qualified calls: io.println, http.get, etc.
	if field, ok := call.Callee.Node.(ast.FieldExpr); ok {
		if ident, ok := field.Object.Node.(ast.IdentExpr); ok {
			method := field.Field.Node
			args := callArgsToGo(call.Args)
			if resolved, ok := resolveQualified(ident.Name, method, args, call); ok {
				return resolved, true
			}
		}
	}

	// Bare calls: env(), len(), keys(), join(), create_agent(), spawn_agents()
	if ident, ok := call.Callee.Node.(ast.IdentExpr); ok {
		args := callArgsToGo(call.Args)
		switch ident.Name {
		case "create_agent":
			return resolveCreateAgent(call), true
		case "spawn_agents":
			return fmt.Sprintf("haira.SpawnAgents(%s)", args), true
		case "env":
			// env("KEY") → haira.Env("KEY")
			// env("KEY", float) → haira.EnvFloat("KEY")
			// env("KEY", int) → haira.EnvInt("KEY")
			if len(call.Args) >= 2 {
				if hint, ok := call.Args[1].Value.Node.(ast.IdentExpr); ok {
					keyArg := ExprToGo(call.Args[0].Value)
					switch hint.Name {
					case "float":
						return fmt.Sprintf("haira.EnvFloat(%s)", keyArg), true
					case "int":
						return fmt.Sprintf("haira.EnvInt(%s)", keyArg), true
					}
				}
			}
			return fmt.Sprintf("haira.Env(%s)", args), true
		case "len":
			return fmt.Sprintf("haira.Len(%s)", args), true
		case "keys":
			return fmt.Sprintf("haira.Keys(%s)", args), true
		case "join":
			return fmt.Sprintf("haira.Join(%s)", args), true
		}
	}

	return "", false
}

// ResolveStdlibMethodCall checks if a method call is a known stdlib call.
func ResolveStdlibMethodCall(mc ast.MethodCallExpr) (string, bool) {
	if ident, ok := mc.Receiver.Node.(ast.IdentExpr); ok {
		method := mc.Method.Node
		args := callArgsToGo(mc.Args)
		stub := ast.CallExpr{Callee: mc.Receiver, Args: mc.Args}
		return resolveQualified(ident.Name, method, args, stub)
	}
	return "", false
}

func resolveQualified(module, method, args string, call ast.CallExpr) (string, bool) {
	switch module {
	case "io":
		switch method {
		case "println":
			return fmt.Sprintf("haira.Println(%s)", args), true
		case "print":
			return fmt.Sprintf("haira.Print(%s)", args), true
		case "printf":
			return fmt.Sprintf("haira.Printf(%s)", args), true
		case "readln":
			return "haira.Readln()", true
		case "read_file":
			return fmt.Sprintf("haira.ReadFile(%s)", args), true
		case "eprintln":
			return fmt.Sprintf("haira.Eprintln(%s)", args), true
		case "eprintf":
			return fmt.Sprintf("haira.Eprintf(%s)", args), true
		}
	case "http":
		switch method {
		case "client":
			return fmt.Sprintf("haira.HttpClient(%s)", args), true
		case "get":
			return fmt.Sprintf("haira.HttpGet(%s)", args), true
		case "get_with_headers":
			return fmt.Sprintf("haira.HttpGetWithHeaders(%s)", args), true
		case "post":
			return fmt.Sprintf("haira.HttpPost(%s)", args), true
		case "post_with_headers":
			return fmt.Sprintf("haira.HttpPostWithHeaders(%s)", args), true
		case "put":
			return fmt.Sprintf("haira.HttpPut(%s)", args), true
		case "put_with_headers":
			return fmt.Sprintf("haira.HttpPutWithHeaders(%s)", args), true
		case "delete":
			return fmt.Sprintf("haira.HttpDelete(%s)", args), true
		case "delete_with_headers":
			return fmt.Sprintf("haira.HttpDeleteWithHeaders(%s)", args), true
		case "encode_uri":
			return fmt.Sprintf("haira.HttpEncodeURI(%s)", args), true
		case "Server":
			return resolveServerCall(call), true
		}
	case "mcp":
		switch method {
		case "Server":
			return resolveMCPServerCall(call), true
		}
	case "vector":
		switch method {
		case "embed":
			return resolveVectorEmbed(call), true
		case "embed_batch":
			return resolveVectorEmbedBatch(call), true
		case "collection":
			return resolveVectorCollection(call), true
		case "insert":
			return fmt.Sprintf("vector.VectorInsert(%s)", args), true
		case "search":
			return fmt.Sprintf("vector.VectorSearch(%s)", args), true
		case "format":
			return fmt.Sprintf("vector.VectorFormat(%s)", args), true
		}
	case "json":
		switch method {
		case "marshal", "encode":
			return fmt.Sprintf("haira.JSONMarshal(%s)", args), true
		case "unmarshal", "decode":
			return fmt.Sprintf("haira.JSONUnmarshal(%s)", args), true
		case "marshal_pretty":
			return fmt.Sprintf("haira.JSONMarshalPretty(%s)", args), true
		case "parse":
			return fmt.Sprintf("haira.JSONParse(%s)", args), true
		}
	case "string":
		switch method {
		case "len":
			return fmt.Sprintf("haira.StringLen(%s)", args), true
		case "is_empty":
			return fmt.Sprintf("haira.StringIsEmpty(%s)", args), true
		case "trim":
			return fmt.Sprintf("haira.StringTrim(%s)", args), true
		case "trim_left":
			return fmt.Sprintf("haira.StringTrimLeft(%s)", args), true
		case "trim_right":
			return fmt.Sprintf("haira.StringTrimRight(%s)", args), true
		case "to_upper":
			return fmt.Sprintf("haira.StringToUpper(%s)", args), true
		case "to_lower":
			return fmt.Sprintf("haira.StringToLower(%s)", args), true
		case "contains":
			return fmt.Sprintf("haira.StringContains(%s)", args), true
		case "starts_with":
			return fmt.Sprintf("haira.StringStartsWith(%s)", args), true
		case "ends_with":
			return fmt.Sprintf("haira.StringEndsWith(%s)", args), true
		case "index_of":
			return fmt.Sprintf("haira.StringIndexOf(%s)", args), true
		case "last_index_of":
			return fmt.Sprintf("haira.StringLastIndexOf(%s)", args), true
		case "split":
			return fmt.Sprintf("haira.StringSplit(%s)", args), true
		case "join":
			return fmt.Sprintf("haira.StringJoin(%s)", args), true
		case "substring":
			return fmt.Sprintf("haira.StringSubstring(%s)", args), true
		case "char_at":
			return fmt.Sprintf("haira.StringCharAt(%s)", args), true
		case "replace":
			return fmt.Sprintf("haira.StringReplace(%s)", args), true
		case "replace_all":
			return fmt.Sprintf("haira.StringReplaceAll(%s)", args), true
		case "repeat":
			return fmt.Sprintf("haira.StringRepeat(%s)", args), true
		case "slugify":
			return fmt.Sprintf("haira.StringSlugify(%s)", args), true
		case "basename":
			return fmt.Sprintf("haira.StringBasename(%s)", args), true
		case "strip_ext":
			return fmt.Sprintf("haira.StringStripExt(%s)", args), true
		case "ext":
			return fmt.Sprintf("haira.StringExt(%s)", args), true
		case "pad_left":
			return fmt.Sprintf("haira.StringPadLeft(%s)", args), true
		case "pad_right":
			return fmt.Sprintf("haira.StringPadRight(%s)", args), true
		case "truncate":
			return fmt.Sprintf("haira.StringTruncate(%s)", args), true
		case "extract_between":
			return fmt.Sprintf("haira.StringExtractBetween(%s)", args), true
		case "capitalize":
			return fmt.Sprintf("haira.StringCapitalize(%s)", args), true
		case "title":
			return fmt.Sprintf("haira.StringTitle(%s)", args), true
		case "reverse":
			return fmt.Sprintf("haira.StringReverse(%s)", args), true
		case "count":
			return fmt.Sprintf("haira.StringCount(%s)", args), true
		case "lines":
			return fmt.Sprintf("haira.StringLines(%s)", args), true
		case "words":
			return fmt.Sprintf("haira.StringWords(%s)", args), true
		case "shell_escape":
			return fmt.Sprintf("haira.StringShellEscape(%s)", args), true
		case "detect_language":
			return fmt.Sprintf("haira.DetectLanguage(%s)", args), true
		}
	case "regex":
		switch method {
		case "is_match":
			return fmt.Sprintf("haira.RegexIsMatch(%s)", args), true
		case "find":
			return fmt.Sprintf("haira.RegexFind(%s)", args), true
		case "find_all":
			return fmt.Sprintf("haira.RegexFindAll(%s)", args), true
		case "replace":
			return fmt.Sprintf("haira.RegexReplace(%s)", args), true
		case "replace_all":
			return fmt.Sprintf("haira.RegexReplaceAll(%s)", args), true
		case "captures":
			return fmt.Sprintf("haira.RegexCaptures(%s)", args), true
		case "split":
			return fmt.Sprintf("haira.RegexSplit(%s)", args), true
		}
	case "math":
		switch method {
		case "abs":
			return fmt.Sprintf("haira.MathAbs(%s)", args), true
		case "min":
			return fmt.Sprintf("haira.MathMin(%s)", args), true
		case "max":
			return fmt.Sprintf("haira.MathMax(%s)", args), true
		case "clamp":
			return fmt.Sprintf("haira.MathClamp(%s)", args), true
		case "floor":
			return fmt.Sprintf("haira.MathFloor(%s)", args), true
		case "ceil":
			return fmt.Sprintf("haira.MathCeil(%s)", args), true
		case "round":
			return fmt.Sprintf("haira.MathRound(%s)", args), true
		case "trunc":
			return fmt.Sprintf("haira.MathTrunc(%s)", args), true
		case "pow":
			return fmt.Sprintf("haira.MathPow(%s)", args), true
		case "sqrt":
			return fmt.Sprintf("haira.MathSqrt(%s)", args), true
		case "cbrt":
			return fmt.Sprintf("haira.MathCbrt(%s)", args), true
		case "exp":
			return fmt.Sprintf("haira.MathExp(%s)", args), true
		case "log":
			return fmt.Sprintf("haira.MathLog(%s)", args), true
		case "log10":
			return fmt.Sprintf("haira.MathLog10(%s)", args), true
		case "log2":
			return fmt.Sprintf("haira.MathLog2(%s)", args), true
		case "sin":
			return fmt.Sprintf("haira.MathSin(%s)", args), true
		case "cos":
			return fmt.Sprintf("haira.MathCos(%s)", args), true
		case "tan":
			return fmt.Sprintf("haira.MathTan(%s)", args), true
		case "asin":
			return fmt.Sprintf("haira.MathAsin(%s)", args), true
		case "acos":
			return fmt.Sprintf("haira.MathAcos(%s)", args), true
		case "atan":
			return fmt.Sprintf("haira.MathAtan(%s)", args), true
		case "atan2":
			return fmt.Sprintf("haira.MathAtan2(%s)", args), true
		case "random":
			return "haira.MathRandom()", true
		case "random_int":
			return fmt.Sprintf("haira.MathRandomInt(%s)", args), true
		case "pi":
			return "haira.MathPI", true
		case "e":
			return "haira.MathE", true
		}
	case "conv":
		switch method {
		case "int_to_string":
			return fmt.Sprintf("haira.ConvIntToString(%s)", args), true
		case "float_to_string":
			return fmt.Sprintf("haira.ConvFloatToString(%s)", args), true
		case "bool_to_string":
			return fmt.Sprintf("haira.ConvBoolToString(%s)", args), true
		case "string_to_int":
			return fmt.Sprintf("haira.ConvStringToInt(%s)", args), true
		case "string_to_float":
			return fmt.Sprintf("haira.ConvStringToFloat(%s)", args), true
		case "string_to_bool":
			return fmt.Sprintf("haira.ConvStringToBool(%s)", args), true
		case "int_to_float":
			return fmt.Sprintf("haira.ConvIntToFloat(%s)", args), true
		case "float_to_int":
			return fmt.Sprintf("haira.ConvFloatToInt(%s)", args), true
		case "int_to_hex":
			return fmt.Sprintf("haira.ConvIntToHex(%s)", args), true
		case "int_to_binary":
			return fmt.Sprintf("haira.ConvIntToBinary(%s)", args), true
		case "int_to_octal":
			return fmt.Sprintf("haira.ConvIntToOctal(%s)", args), true
		case "hex_to_int":
			return fmt.Sprintf("haira.ConvHexToInt(%s)", args), true
		case "to_string":
			return fmt.Sprintf("haira.ConvToString(%s)", args), true
		}
	case "array":
		switch method {
		case "len":
			return fmt.Sprintf("haira.ArrayLen(%s)", args), true
		case "is_empty":
			return fmt.Sprintf("haira.ArrayIsEmpty(%s)", args), true
		case "first":
			return fmt.Sprintf("haira.ArrayFirst(%s)", args), true
		case "last":
			return fmt.Sprintf("haira.ArrayLast(%s)", args), true
		case "get":
			return fmt.Sprintf("haira.ArrayGet(%s)", args), true
		case "push":
			return fmt.Sprintf("haira.ArrayPush(%s)", args), true
		case "pop":
			return fmt.Sprintf("haira.ArrayPop(%s)", args), true
		case "insert":
			return fmt.Sprintf("haira.ArrayInsert(%s)", args), true
		case "remove":
			return fmt.Sprintf("haira.ArrayRemove(%s)", args), true
		case "slice":
			return fmt.Sprintf("haira.ArraySlice(%s)", args), true
		case "take":
			return fmt.Sprintf("haira.ArrayTake(%s)", args), true
		case "drop":
			return fmt.Sprintf("haira.ArrayDrop(%s)", args), true
		case "contains":
			return fmt.Sprintf("haira.ArrayContains(%s)", args), true
		case "index_of":
			return fmt.Sprintf("haira.ArrayIndexOf(%s)", args), true
		case "reverse":
			return fmt.Sprintf("haira.ArrayReverse(%s)", args), true
		case "concat":
			return fmt.Sprintf("haira.ArrayConcat(%s)", args), true
		case "flatten":
			return fmt.Sprintf("haira.ArrayFlatten(%s)", args), true
		case "unique":
			return fmt.Sprintf("haira.ArrayUnique(%s)", args), true
		case "sort":
			return fmt.Sprintf("haira.ArraySort(%s)", args), true
		case "join":
			return fmt.Sprintf("haira.ArrayJoin(%s)", args), true
		case "map":
			return fmt.Sprintf("haira.ArrayMap(%s)", args), true
		case "filter":
			return fmt.Sprintf("haira.ArrayFilter(%s)", args), true
		case "reduce":
			return fmt.Sprintf("haira.ArrayReduce(%s)", args), true
		case "find":
			return fmt.Sprintf("haira.ArrayFind(%s)", args), true
		case "find_index":
			return fmt.Sprintf("haira.ArrayFindIndex(%s)", args), true
		case "sort_by":
			return fmt.Sprintf("haira.ArraySortBy(%s)", args), true
		case "every":
			return fmt.Sprintf("haira.ArrayEvery(%s)", args), true
		case "some":
			return fmt.Sprintf("haira.ArraySome(%s)", args), true
		case "for_each":
			return fmt.Sprintf("haira.ArrayForEach(%s)", args), true
		case "flat_map":
			return fmt.Sprintf("haira.ArrayFlatMap(%s)", args), true
		}
	case "map":
		switch method {
		case "len":
			return fmt.Sprintf("haira.MapLen(%s)", args), true
		case "is_empty":
			return fmt.Sprintf("haira.MapIsEmpty(%s)", args), true
		case "get":
			return fmt.Sprintf("haira.MapGet(%s)", args), true
		case "has":
			return fmt.Sprintf("haira.MapHas(%s)", args), true
		case "set":
			return fmt.Sprintf("haira.MapSet(%s)", args), true
		case "remove":
			return fmt.Sprintf("haira.MapRemove(%s)", args), true
		case "keys":
			return fmt.Sprintf("haira.MapKeys(%s)", args), true
		case "values":
			return fmt.Sprintf("haira.MapValues(%s)", args), true
		case "entries":
			return fmt.Sprintf("haira.MapEntries(%s)", args), true
		case "merge":
			return fmt.Sprintf("haira.MapMerge(%s)", args), true
		case "contains_value":
			return fmt.Sprintf("haira.MapContainsValue(%s)", args), true
		}
	case "postgres":
		switch method {
		case "connect":
			return fmt.Sprintf("postgres.PostgresConnect(%s)", args), true
		case "generate_upsert":
			// 2-arg: postgres.generate_upsert(tables, schema)
			// 3-arg: postgres.generate_upsert(tables, schema, conflicts)
			// These functions live in the excel package (they access ExcelTables internals)
			if len(call.Args) >= 3 {
				return fmt.Sprintf("excel.PostgresGenerateUpsertWithConflicts(%s)", args), true
			}
			return fmt.Sprintf("excel.PostgresGenerateUpsert(%s)", args), true
		case "escape":
			return fmt.Sprintf("postgres.PostgresEscape(%s)", args), true
		case "quote_identifier":
			return fmt.Sprintf("postgres.QuoteIdentifier(%s)", args), true
		}
	case "slack":
		switch method {
		case "send":
			return fmt.Sprintf("slack.SlackSend(%s)", args), true
		case "send_blocks":
			return fmt.Sprintf("slack.SlackSendBlocks(%s)", args), true
		case "header":
			return fmt.Sprintf("slack.SlackHeader(%s)", args), true
		case "section":
			return fmt.Sprintf("slack.SlackSection(%s)", args), true
		case "divider":
			return "slack.SlackDivider()", true
		case "context":
			return fmt.Sprintf("slack.SlackContext(%s)", args), true
		case "client":
			return fmt.Sprintf("slack.SlackNewClient(%s)", args), true
		case "send_alert":
			return fmt.Sprintf("slack.SlackSendAlert(%s)", args), true
		}
	case "excel":
		switch method {
		case "open":
			return fmt.Sprintf("excel.ExcelOpen(%s)", args), true
		case "read_sheets":
			return fmt.Sprintf("excel.ExcelReadSheets(%s)", args), true
		case "read_config":
			return fmt.Sprintf("excel.ExcelReadConfig(%s)", args), true
		}
	case "log":
		// log.info/warn/error/render inside steps → haira.StepLog/StepRender with injected context
		if activeWorkflowName != "" && activeStepName != "" {
			switch method {
			case "info":
				return fmt.Sprintf("haira.StepLog(%q, %q, \"info\", %s)", activeWorkflowName, activeStepName, args), true
			case "warn":
				return fmt.Sprintf("haira.StepLog(%q, %q, \"warn\", %s)", activeWorkflowName, activeStepName, args), true
			case "error":
				return fmt.Sprintf("haira.StepLog(%q, %q, \"error\", %s)", activeWorkflowName, activeStepName, args), true
			case "render":
				return fmt.Sprintf("haira.StepRender(%q, %q, %s)", activeWorkflowName, activeStepName, args), true
			}
		} else {
			// Outside steps: print with level prefix to stdout/stderr
			switch method {
			case "info":
				return fmt.Sprintf("haira.LogPrint(\"info\", %s)", args), true
			case "warn":
				return fmt.Sprintf("haira.LogPrint(\"warn\", %s)", args), true
			case "error":
				return fmt.Sprintf("haira.LogPrint(\"error\", %s)", args), true
			}
		}
	case "time":
		switch method {
		case "sleep":
			return fmt.Sprintf("haira.TimeSleep(%s)", args), true
		case "now":
			return "haira.TimeNow()", true
		case "format":
			return fmt.Sprintf("haira.TimeFormat(%s)", args), true
		case "parse":
			return fmt.Sprintf("haira.TimeParse(%s)", args), true
		case "since":
			return fmt.Sprintf("haira.TimeSince(%s)", args), true
		case "after":
			return fmt.Sprintf("haira.TimeAfter(%s)", args), true
		case "tick":
			return fmt.Sprintf("haira.TimeTick(%s)", args), true
		case "slug":
			return "haira.TimeSlug()", true
		case "timestamp":
			return "haira.TimeTimestamp()", true
		}
	case "os":
		switch method {
		case "cwd":
			return "haira.OsCwd()", true
		case "exec":
			return fmt.Sprintf("haira.OsExec(%s)", args), true
		case "exec_timeout":
			return fmt.Sprintf("haira.OsExecTimeout(%s)", args), true
		case "safe_exec":
			return fmt.Sprintf("haira.OsSafeExec(%s)", args), true
		case "arch":
			return "haira.OsArch()", true
		case "platform":
			return "haira.OsPlatform()", true
		case "hostname":
			return "haira.OsHostname()", true
		case "exit":
			return fmt.Sprintf("haira.OsExit(%s)", args), true
		case "args":
			return "haira.OsArgs()", true
		case "getenv":
			return fmt.Sprintf("haira.OsGetenv(%s)", args), true
		case "setenv":
			return fmt.Sprintf("haira.OsSetenv(%s)", args), true
		case "environ":
			return "haira.OsEnviron()", true
		case "chdir":
			return fmt.Sprintf("haira.OsChdir(%s)", args), true
		case "temp_dir":
			return "haira.OsTempDir()", true
		}
	case "fs":
		switch method {
		case "read_file":
			return fmt.Sprintf("haira.FsReadFile(%s)", args), true
		case "write_file":
			return fmt.Sprintf("haira.FsWriteFile(%s)", args), true
		case "append_file":
			return fmt.Sprintf("haira.FsAppendFile(%s)", args), true
		case "exists":
			return fmt.Sprintf("haira.FsExists(%s)", args), true
		case "remove":
			return fmt.Sprintf("haira.FsRemove(%s)", args), true
		case "remove_all":
			return fmt.Sprintf("haira.FsRemoveAll(%s)", args), true
		case "rename":
			return fmt.Sprintf("haira.FsRename(%s)", args), true
		case "copy":
			return fmt.Sprintf("haira.FsCopy(%s)", args), true
		case "mkdir":
			return fmt.Sprintf("haira.FsMkdir(%s)", args), true
		case "mkdir_all":
			return fmt.Sprintf("haira.FsMkdirAll(%s)", args), true
		case "read_dir":
			return fmt.Sprintf("haira.FsReadDir(%s)", args), true
		case "stat":
			return fmt.Sprintf("haira.FsStat(%s)", args), true
		}
	case "observe":
		switch method {
		case "usage":
			return "haira.ObserveUsage()", true
		case "agent_usage":
			return fmt.Sprintf("haira.ObserveAgentUsage(%s)", args), true
		case "session_usage":
			return fmt.Sprintf("haira.ObserveSessionUsage(%s)", args), true
		case "model_usage":
			return fmt.Sprintf("haira.ObserveModelUsage(%s)", args), true
		case "events":
			return "haira.ObserveEvents()", true
		case "agent_events":
			return fmt.Sprintf("haira.ObserveAgentEvents(%s)", args), true
		case "reset":
			return "haira.ObserveReset()", true
		case "start":
			return fmt.Sprintf("haira.ObserveStart(%s)", args), true
		case "cost":
			return "haira.ObserveCost()", true
		case "agent_cost":
			return fmt.Sprintf("haira.ObserveAgentCost(%s)", args), true
		case "use":
			return fmt.Sprintf("haira.ObserveExport(%s)", args), true
		}
	case "store":
		switch method {
		case "database":
			return fmt.Sprintf("haira.SetStoreURL(%s)", args), true
		}
	case "auth":
		switch method {
		case "login":
			return fmt.Sprintf("auth.AuthLogin(%s)", args), true
		case "logout":
			return "auth.AuthLogout()", true
		case "status":
			return "auth.AuthStatus()", true
		case "resolve_token":
			return fmt.Sprintf("auth.ResolveToken(%s)", args), true
		}
	case "langfuse":
		switch method {
		case "exporter":
			return fmt.Sprintf("langfuse.LangfuseExporter(%s)", args), true
		}
	case "gitlab":
		switch method {
		case "Client":
			return resolveClientConstructor("gitlab.GitlabNewClient", call), true
		case "client":
			return resolveClientConstructor("gitlab.GitlabConnect", call), true
		}
	case "github":
		switch method {
		case "Client":
			return resolveClientConstructor("github.GithubNewClient", call), true
		case "client":
			return resolveClientConstructor("github.GithubConnect", call), true
		}
	case "agent":
		// Note: "agent" is also a keyword, so agent.create() can only work
		// via import alias: `import rt from "agent"` → rt.create(...)
		switch method {
		case "create":
			return resolveCreateAgent(call), true
		case "spawn":
			return fmt.Sprintf("haira.SpawnAgents(%s)", args), true
		}
	case "websearch":
		switch method {
		case "search":
			return fmt.Sprintf("websearch.DuckDuckGoSearch(%s)", args), true
		case "fetch":
			return fmt.Sprintf("websearch.WebFetch(%s)", args), true
		}
	case "healthcheck":
		switch method {
		case "check":
			return fmt.Sprintf("healthcheck.Check(%s)", args), true
		case "check_all":
			return fmt.Sprintf("healthcheck.CheckAll(%s)", args), true
		}
	case "algolia":
		switch method {
		case "client":
			return fmt.Sprintf("algolia.AlgoliaNewClient(%s)", args), true
		case "hits":
			return fmt.Sprintf("algolia.AlgoliaHits(%s)", args), true
		case "hits_to_table":
			return fmt.Sprintf("algolia.AlgoliaHitsToTable(%s)", args), true
		case "hits_to_product_cards":
			return fmt.Sprintf("algolia.AlgoliaHitsToProductCards(%s)", args), true
		}
	case "meilisearch":
		switch method {
		case "client":
			return fmt.Sprintf("meilisearch.MeilisearchNewClient(%s)", args), true
		case "hits":
			return fmt.Sprintf("meilisearch.MeilisearchHits(%s)", args), true
		case "hits_to_table":
			return fmt.Sprintf("meilisearch.MeilisearchHitsToTable(%s)", args), true
		case "hits_to_product_cards":
			return fmt.Sprintf("meilisearch.MeilisearchHitsToProductCards(%s)", args), true
		}
	case "ui":
		switch method {
		case "key_value":
			return fmt.Sprintf("haira.UiNewKeyValue(%s)", args), true
		case "status_card":
			return fmt.Sprintf("haira.UiNewStatusCard(%s)", args), true
		case "group":
			return fmt.Sprintf("haira.UiNewGroup(%s)", args), true
		case "confirm":
			return fmt.Sprintf("haira.UiNewConfirm(%s)", args), true
		case "chart":
			return fmt.Sprintf("haira.UiNewChart(%s)", args), true
		case "table":
			return fmt.Sprintf("haira.UiNewTable(%s)", args), true
		case "product_cards":
			return fmt.Sprintf("haira.UiNewProductCards(%s)", args), true
		case "markdown":
			return fmt.Sprintf("haira.UiNewMarkdown(%s)", args), true
		case "code_block":
			return fmt.Sprintf("haira.UiNewCodeBlock(%s)", args), true
		case "diff":
			return fmt.Sprintf("haira.UiNewDiff(%s)", args), true
		case "progress":
			return fmt.Sprintf("haira.UiNewProgress(%s)", args), true
		case "choices":
			return fmt.Sprintf("haira.UiNewChoices(%s)", args), true
		}
	}
	return "", false
}

// resolveClientConstructor generates a Go call with positional args + named args collected into a map.
// e.g., gitlab.Client(token, project: 359) → haira.GitlabNewClient(token, map[string]any{"project": 359})
func resolveClientConstructor(goFunc string, call ast.CallExpr) string {
	var positional []string
	var named []string
	for _, arg := range call.Args {
		if arg.Name != nil {
			named = append(named, fmt.Sprintf("%q: %s", arg.Name.Node, ExprToGo(arg.Value)))
		} else {
			positional = append(positional, ExprToGo(arg.Value))
		}
	}
	parts := positional
	if len(named) > 0 {
		parts = append(parts, fmt.Sprintf("map[string]any{%s}", strings.Join(named, ", ")))
	} else {
		parts = append(parts, "nil")
	}
	return fmt.Sprintf("%s(%s)", goFunc, strings.Join(parts, ", "))
}

func resolveServerCall(call ast.CallExpr) string {
	if len(call.Args) == 1 {
		if list, ok := call.Args[0].Value.Node.(ast.ListExpr); ok {
			refs := make([]string, len(list.Elems))
			for i, item := range list.Elems {
				if ident, ok := item.Node.(ast.IdentExpr); ok {
					refs[i] = "workflowDef" + SnakeToPascal(ident.Name)
				} else {
					refs[i] = ExprToGo(item)
				}
			}
			return fmt.Sprintf("haira.NewServer([]*haira.WorkflowDef{%s})", strings.Join(refs, ", "))
		}
	}
	args := callArgsToGo(call.Args)
	return fmt.Sprintf("haira.NewServer(%s)", args)
}

func resolveMCPServerCall(call ast.CallExpr) string {
	if len(call.Args) == 1 {
		if list, ok := call.Args[0].Value.Node.(ast.ListExpr); ok {
			refs := make([]string, len(list.Elems))
			for i, item := range list.Elems {
				if ident, ok := item.Node.(ast.IdentExpr); ok {
					refs[i] = "workflowDef" + SnakeToPascal(ident.Name)
				} else {
					refs[i] = ExprToGo(item)
				}
			}
			return fmt.Sprintf("haira.NewMCPServer([]*haira.WorkflowDef{%s})", strings.Join(refs, ", "))
		}
	}
	args := callArgsToGo(call.Args)
	return fmt.Sprintf("haira.NewMCPServer(%s)", args)
}

func callArgsToGo(args []ast.Argument) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = ExprToGo(a.Value)
	}
	return strings.Join(parts, ", ")
}

// resolveVectorEmbed converts vector.embed(provider, text) → vector.VectorEmbed(providerVar, text)
func resolveVectorEmbed(call ast.CallExpr) string {
	if len(call.Args) < 2 {
		return "vector.VectorEmbed(nil, \"\")"
	}
	providerArg := resolveProviderArg(call.Args[0].Value)
	textArg := ExprToGo(call.Args[1].Value)
	return fmt.Sprintf("vector.VectorEmbed(%s, %s)", providerArg, textArg)
}

// resolveVectorEmbedBatch converts vector.embed_batch(provider, texts) → vector.VectorEmbedBatch(providerVar, texts)
func resolveVectorEmbedBatch(call ast.CallExpr) string {
	if len(call.Args) < 2 {
		return "vector.VectorEmbedBatch(nil, nil)"
	}
	providerArg := resolveProviderArg(call.Args[0].Value)
	textsArg := ExprToGo(call.Args[1].Value)
	return fmt.Sprintf("vector.VectorEmbedBatch(%s, %s)", providerArg, textsArg)
}

// resolveVectorCollection converts:
//
//	vector.collection(db, "name", dimensions: N) → vector.VectorNewCollection(db, "name", N)
//	vector.collection("name", dimensions: N)     → vector.VectorNewCollection(nil, "name", N)
func resolveVectorCollection(call ast.CallExpr) string {
	if len(call.Args) < 1 {
		return "vector.VectorNewCollection(nil, \"\", 0)"
	}

	// Detect if first arg is a string literal (no-db form)
	firstIsString := false
	if lit, ok := call.Args[0].Value.Node.(ast.LiteralExpr); ok {
		if _, ok := lit.Lit.(ast.StringLit); ok {
			firstIsString = true
		}
	}

	var dbArg, nameArg string
	var restArgs []ast.Argument

	if firstIsString {
		// No db: vector.collection("name", dimensions: N)
		dbArg = "nil"
		nameArg = ExprToGo(call.Args[0].Value)
		restArgs = call.Args[1:]
	} else {
		// With db: vector.collection(db, "name", dimensions: N)
		if len(call.Args) < 2 {
			return "vector.VectorNewCollection(nil, \"\", 0)"
		}
		dbArg = ExprToGo(call.Args[0].Value)
		nameArg = ExprToGo(call.Args[1].Value)
		restArgs = call.Args[2:]
	}

	// Look for named arg "dimensions"
	dimArg := "0"
	for _, arg := range restArgs {
		if arg.Name != nil && arg.Name.Node == "dimensions" {
			dimArg = ExprToGo(arg.Value)
			break
		}
	}
	return fmt.Sprintf("vector.VectorNewCollection(%s, %s, %s)", dbArg, nameArg, dimArg)
}

// resolveProviderArg resolves a provider identifier to its Go variable name.
// e.g., "openai_embed" → "providerOpenaiEmbed"
func resolveProviderArg(expr ast.Expr) string {
	if ident, ok := expr.Node.(ast.IdentExpr); ok {
		return goVarName("provider", ident.Name)
	}
	return ExprToGo(expr)
}

// resolveCreateAgent generates haira.CreateAgent(config, provider, tools) with
// proper provider variable resolution for the second argument.
func resolveCreateAgent(call ast.CallExpr) string {
	parts := make([]string, len(call.Args))
	for i, a := range call.Args {
		if i == 1 {
			// Second arg is the provider — resolve to providerXxx variable
			parts[i] = resolveProviderArg(a.Value)
		} else {
			parts[i] = ExprToGo(a.Value)
		}
	}
	return fmt.Sprintf("haira.CreateAgent(%s)", strings.Join(parts, ", "))
}

// IsStdlibImport returns whether a Haira import path maps to a stdlib module.
func IsStdlibImport(path string) bool {
	switch path {
	case "io", "http", "mcp", "env", "json", "postgres", "slack", "excel", "time",
		"string", "regex", "math", "conv", "array", "map", "log", "ui", "vector",
		"observe", "fs", "os", "gitlab", "github", "langfuse", "algolia", "meilisearch",
		"auth", "agent", "websearch", "healthcheck":
		return true
	}
	return false
}
