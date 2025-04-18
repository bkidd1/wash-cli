// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import fetch from 'node-fetch';

const API_URL = 'http://localhost:3000/';

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {
	// Command to analyze coding pathways for current file
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
			const fileName = document.fileName;

			await analyzeCode([{ fileName, content: text }]);
		} catch (error) {
			vscode.window.showErrorMessage(`Error: ${error}`);
		}
	});

	// Command to analyze multiple files
	const analyzeMultipleFiles = vscode.commands.registerCommand('wash.analyzeMultipleFiles', async () => {
		try {
			// Get workspace folders
			const workspaceFolders = vscode.workspace.workspaceFolders;
			if (!workspaceFolders) {
				vscode.window.showErrorMessage('No workspace folder found!');
				return;
			}

			// Show file picker with default path to workspace
			const files = await vscode.window.showOpenDialog({
				canSelectMany: true,
				openLabel: 'Analyze Selected Files',
				defaultUri: workspaceFolders[0].uri,
				filters: {
					'All files': ['*']
				},
				title: 'Select files to analyze'
			});

			if (!files || files.length === 0) {
				return;
			}

			// Read all selected files
			const fileContents = await Promise.all(
				files.map(async (file) => {
					const document = await vscode.workspace.openTextDocument(file);
					return {
						fileName: file.fsPath,
						content: document.getText()
					};
				})
			);

			await analyzeCode(fileContents);
		} catch (error) {
			vscode.window.showErrorMessage(`Error: ${error}`);
		}
	});

	// Command to explore project structure
	const exploreProject = vscode.commands.registerCommand('wash.exploreProject', async () => {
		try {
			const workspaceFolders = vscode.workspace.workspaceFolders;
			if (!workspaceFolders) {
				vscode.window.showErrorMessage('No workspace folder found!');
				return;
			}

			const projectStructure = await getProjectStructure(workspaceFolders[0].uri.fsPath);
			await analyzeProjectStructure(projectStructure);
		} catch (error) {
			vscode.window.showErrorMessage(`Error exploring project: ${error}`);
		}
	});

	context.subscriptions.push(analyzePathways, analyzeMultipleFiles, exploreProject);
}

async function analyzeCode(files: { fileName: string; content: string }[]) {
	// Create a progress indicator
	await vscode.window.withProgress({
		location: vscode.ProgressLocation.Notification,
		title: "Analyzing coding pathways...",
		cancellable: false
	}, async (progress) => {
		try {
			let totalAnalysis = '';
			const totalFiles = files.length;

			// First analyze project structure if multiple files
			if (files.length > 1) {
				const workspaceFolders = vscode.workspace.workspaceFolders;
				if (workspaceFolders) {
					progress.report({ message: 'Analyzing project structure...' });
					const projectStructure = await getProjectStructure(workspaceFolders[0].uri.fsPath);
					const structureResponse = await fetch(`${API_URL}analyze-structure`, {
						method: 'POST',
						headers: {
							'Content-Type': 'application/json',
						},
						body: JSON.stringify({ projectStructure })
					});

					if (!structureResponse.ok) {
						const errorText = await structureResponse.text();
						throw new Error(`Server responded with status ${structureResponse.status}: ${errorText}`);
					}

					const structureData = await structureResponse.json() as { analysis: string };
					totalAnalysis += `=== Project Structure Analysis ===\n\n${structureData.analysis}\n\n`;
				}
			}

			// Analyze each file individually
			for (let i = 0; i < files.length; i++) {
				const file = files[i];
				progress.report({ 
					message: `Analyzing file ${i + 1} of ${totalFiles}: ${file.fileName}`,
					increment: (100 / totalFiles)
				});

				const response = await fetch(`${API_URL}analyze-file`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify({
						fileName: file.fileName,
						content: file.content
					})
				});

				if (!response.ok) {
					const errorText = await response.text();
					throw new Error(`Server responded with status ${response.status}: ${errorText}`);
				}

				const data = await response.json() as { fileName: string; analysis: string };
				totalAnalysis += `=== Analysis for ${data.fileName} ===\n\n${data.analysis}\n\n`;
			}

			// Display the combined analysis in a new webview
			const panel = vscode.window.createWebviewPanel(
				'washAnalysis',
				'Wash Analysis',
				vscode.ViewColumn.One,
				{
					enableScripts: true
				}
			);

			panel.webview.html = getWebviewContent(totalAnalysis);
		} catch (error) {
			vscode.window.showErrorMessage(`Error analyzing code: ${error}`);
		}
	});
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

async function getProjectStructure(rootPath: string, depth: number = 2): Promise<{ path: string; type: 'file' | 'directory'; children?: any[] }> {
	try {
		console.log('Getting project structure for path:', rootPath, 'at depth:', depth);
		const stats = await fs.promises.stat(rootPath);
		
		if (stats.isFile()) {
			console.log('Found file:', rootPath);
			// For files, only include the path and type
			return {
				path: rootPath,
				type: 'file'
			};
		}

		// If we've reached the maximum depth, return just the directory info
		if (depth <= 0) {
			return {
				path: rootPath,
				type: 'directory'
			};
		}

		console.log('Reading directory:', rootPath);
		const children = await fs.promises.readdir(rootPath);
		console.log('Found children:', children);

		const childPromises = children.map(async (child) => {
			const childPath = path.join(rootPath, child);
			
			// Skip common directories and files that aren't relevant
			const skipPatterns = [
				'node_modules',
				'.git',
				'dist',
				'out',
				'build',
				'coverage',
				'.next',
				'.vscode',
				'.idea',
				'*.log',
				'*.lock',
				'*.map',
				'*.min.js',
				'*.min.css'
			];

			if (skipPatterns.some(pattern => {
				if (pattern.startsWith('*')) {
					return child.endsWith(pattern.slice(1));
				}
				return child === pattern;
			})) {
				console.log('Skipping:', childPath);
				return null;
			}

			return getProjectStructure(childPath, depth - 1);
		});

		const childResults = (await Promise.all(childPromises)).filter(child => child !== null);
		console.log('Processed children for:', rootPath, childResults);
		
		return {
			path: rootPath,
			type: 'directory',
			children: childResults
		};
	} catch (error) {
		console.error('Error in getProjectStructure:', error);
		throw error;
	}
}

async function analyzeProjectStructure(structure: any) {
	await vscode.window.withProgress({
		location: vscode.ProgressLocation.Notification,
		title: "Analyzing project structure...",
		cancellable: false
	}, async (progress) => {
		try {
			console.log('Sending project structure to server:', JSON.stringify(structure, null, 2));
			const response = await fetch(`${API_URL}analyze-structure`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ projectStructure: structure })
			});

			if (!response.ok) {
				const errorText = await response.text();
				console.error('Server error response:', {
					status: response.status,
					statusText: response.statusText,
					errorText
				});
				throw new Error(`Server responded with status ${response.status}: ${errorText}`);
			}

			const data = await response.json() as { analysis: string };
			console.log('Received analysis from server');

			// Display the analysis in a new webview
			const panel = vscode.window.createWebviewPanel(
				'washAnalysis',
				'Wash Project Analysis',
				vscode.ViewColumn.One,
				{
					enableScripts: true
				}
			);

			panel.webview.html = getWebviewContent(data.analysis);
		} catch (error) {
			console.error('Error in analyzeProjectStructure:', error);
			vscode.window.showErrorMessage(`Error analyzing project structure: ${error}`);
		}
	});
}

// This method is called when your extension is deactivated
export function deactivate() {}
