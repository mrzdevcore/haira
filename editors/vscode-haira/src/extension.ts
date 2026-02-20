import * as path from "path";
import { workspace, ExtensionContext, window } from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from "vscode-languageclient/node";

let client: LanguageClient | undefined;

export function activate(context: ExtensionContext) {
  const serverCommand =
    workspace.getConfiguration("haira").get<string>("serverPath") || "haira";
  const serverArgs = ["lsp"];

  const serverOptions: ServerOptions = {
    command: serverCommand,
    args: serverArgs,
    transport: TransportKind.stdio,
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "haira" }],
    synchronize: {
      fileEvents: workspace.createFileSystemWatcher("**/*.haira"),
    },
  };

  client = new LanguageClient(
    "haira-lsp",
    "Haira Language Server",
    serverOptions,
    clientOptions
  );

  client.start().catch((err) => {
    window.showErrorMessage(
      `Failed to start Haira language server: ${err.message}. ` +
        `Ensure the 'haira' binary is installed and on your PATH.`
    );
  });
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}
