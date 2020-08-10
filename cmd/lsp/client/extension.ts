import { workspace, window, ExtensionContext } from 'vscode';
import * as child_process from 'child_process';
import {
	LanguageClient,
	LanguageClientOptions,
} from 'vscode-languageclient';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
	// Options to control the language client
	let clientOptions: LanguageClientOptions = {
		documentSelector: [{ scheme: 'file', language: 'yaml' }],
		outputChannelName: 'crosspls',
	};

	// Create the language client and start the client.
	client = new LanguageClient(
		'languageServerExample',
		'Language Server Example',
		spawnServer,
		clientOptions
	);

	// Start the client. This will also launch the server
	client.start();
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}

async function spawnServer(): Promise<child_process.ChildProcess> {
	let serverProcess = child_process.spawn('./code/github.com/hasheddan/crank/cmd/lsp/server/server')
    serverProcess.on('error', (err: { code?: string; message: string }) => {
		window.showWarningMessage(
			`Failed to spawn crosspls: \`${err.message}\``
		)
	})
    return serverProcess
}