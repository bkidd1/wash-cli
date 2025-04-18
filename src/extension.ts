// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';

const API_URL = 'http://localhost:3000/';

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {
	// Command to analyze coding pathways
	const analyzePathways = vscode.commands.registerCommand('wash.analyzePathways', async () => {
		try {
			// Get the active text editor
			const editor = vscode.window.activeTextEditor;
			if (!editor) {
				vscode.window.showErrorMessage('No active editor found!');
				return;
			}

			// Get the current document content
			const document = editor.document;
			const text = document.getText();

			// Create a progress indicator
			await vscode.window.withProgress({
				location: vscode.ProgressLocation.Notification,
				title: "Analyzing coding pathways...",
				cancellable: false
			}, async (progress) => {
				try {
					console.log('Sending request to:', `${API_URL}analyze`);
					const response = await fetch(`${API_URL}analyze`, {
						method: 'POST',
						headers: {
							'Content-Type': 'application/json',
						},
						body: JSON.stringify({ code: text })
					});

					if (!response.ok) {
						const errorText = await response.text();
						throw new Error(`Server responded with status ${response.status}: ${errorText}`);
					}

					const data = await response.json() as { analysis: string };

					// Display the analysis in a new webview
					const panel = vscode.window.createWebviewPanel(
						'washAnalysis',
						'Wash Analysis',
						vscode.ViewColumn.One,
						{
							enableScripts: true
						}
					);

					panel.webview.html = getWebviewContent(data.analysis);
				} catch (error) {
					vscode.window.showErrorMessage(`Error analyzing code: ${error}`);
				}
			});
		} catch (error) {
			vscode.window.showErrorMessage(`Error: ${error}`);
		}
	});

	context.subscriptions.push(analyzePathways);
}

function getWebviewContent(analysis: string): string {
	return `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Wash Analysis</title>
		<style>
			body {
				padding: 20px;
				font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
				line-height: 1.6;
			}
			.analysis {
				background-color: #f5f5f5;
				padding: 20px;
				border-radius: 8px;
				white-space: pre-wrap;
			}
		</style>
	</head>
	<body>
		<h1>Wash Analysis</h1>
		<div class="analysis">
			${analysis}
		</div>
	</body>
	</html>`;
}

// This method is called when your extension is deactivated
export function deactivate() {}
