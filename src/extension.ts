// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import fetch from 'node-fetch';
import * as os from 'os';
import * as child_process from 'child_process';

const API_URL = 'http://localhost:3000/';
const MEETING_NOTES_FOLDER = '.wash-notes';
const CHAT_MONITOR_INTERVAL = 30000; // 30 seconds

// Add interface for chat monitoring state
interface ChatMonitorState {
	isMonitoring: boolean;
	lastScreenshotTime: number;
	continuousNotesPath: string;
}

// Initialize chat monitoring state
let chatMonitorState: ChatMonitorState = {
	isMonitoring: false,
	lastScreenshotTime: 0,
	continuousNotesPath: ''
};

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

	// Add commands for continuous chat monitoring
	const startMonitoring = vscode.commands.registerCommand('wash.startChatMonitoring', startChatMonitoring);
	const stopMonitoring = vscode.commands.registerCommand('wash.stopChatMonitoring', stopChatMonitoring);
	const generateSummary = vscode.commands.registerCommand('wash.generateSummary', generateChatSummary);

	context.subscriptions.push(
		analyzePathways, 
		analyzeMultipleFiles, 
		exploreProject,
		startMonitoring,
		stopMonitoring,
		generateSummary
	);
}

// Function to ensure the meeting notes folder exists
async function ensureMeetingNotesFolder(): Promise<string> {
	const workspaceFolders = vscode.workspace.workspaceFolders;
	if (!workspaceFolders) {
		throw new Error('No workspace folder found');
	}

	const notesPath = path.join(workspaceFolders[0].uri.fsPath, MEETING_NOTES_FOLDER);
	try {
		await fs.promises.mkdir(notesPath, { recursive: true });
		return notesPath;
	} catch (error) {
		throw new Error(`Failed to create meeting notes folder: ${error}`);
	}
}

// Function to save analysis as meeting notes
async function saveMeetingNotes(type: 'code' | 'chat' | 'structure', content: string): Promise<string> {
	const notesPath = await ensureMeetingNotesFolder();
	const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
	const fileName = `${type}-analysis-${timestamp}.md`;
	const filePath = path.join(notesPath, fileName);

	// Format the content as meeting notes
	const formattedContent = `# Wash Meeting Notes - ${type.charAt(0).toUpperCase() + type.slice(1)} Analysis
*Generated on ${new Date().toLocaleString()}*

## Key Insights
${content}

## Action Items
- [ ] Review and implement suggested improvements
- [ ] Consider alternative approaches discussed
- [ ] Document any successful strategies for future reference

## Next Steps
- [ ] Follow up on identified issues
- [ ] Implement recommended changes
- [ ] Schedule next review if needed

---

*Note: These meeting notes are automatically generated by Wash to help track progress and maintain context between you and Cursor AI.*
`;

	await fs.promises.writeFile(filePath, formattedContent);
	return filePath;
}

// Update the analyzeCode function to save meeting notes
async function analyzeCode(files: { fileName: string; content: string }[]) {
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
					
					// Save structure analysis as meeting notes
					const structureNotesPath = await saveMeetingNotes('structure', structureData.analysis);
					vscode.window.showInformationMessage(`Structure analysis saved to ${structureNotesPath}`);
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
				
				// Save file analysis as meeting notes
				const fileNotesPath = await saveMeetingNotes('code', data.analysis);
				vscode.window.showInformationMessage(`Analysis for ${path.basename(data.fileName)} saved to ${fileNotesPath}`);
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

// Function to capture screenshot of the chat interface
async function captureScreenshot(): Promise<string | null> {
	try {
		// Create a temporary file for the screenshot
		const tempDir = os.tmpdir();
		const screenshotPath = path.join(tempDir, 'chat-screenshot.png');

		// Use screencapture command on macOS
		if (process.platform === 'darwin') {
			child_process.execSync(`screencapture -R 0,0,800,600 ${screenshotPath}`);
		} else {
			// For other platforms, we'll need to implement platform-specific screenshot capture
			vscode.window.showErrorMessage('Screenshot capture is currently only supported on macOS');
			return null;
		}

		// Convert the screenshot to base64
		const screenshotBuffer = await fs.promises.readFile(screenshotPath);
		const base64Screenshot = screenshotBuffer.toString('base64');
		
		// Clean up the temporary file
		await fs.promises.unlink(screenshotPath);

		return `data:image/png;base64,${base64Screenshot}`;
	} catch (error) {
		console.error('Error capturing screenshot:', error);
		return null;
	}
}

// Update the analyzeChatScreenshot function to save meeting notes
async function analyzeChatScreenshot(screenshot: string) {
	await vscode.window.withProgress({
		location: vscode.ProgressLocation.Notification,
		title: "Analyzing chat history...",
		cancellable: false
	}, async (progress) => {
		try {
			const response = await fetch(`${API_URL}analyze-chat`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ screenshot })
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(`Server responded with status ${response.status}: ${errorText}`);
			}

			const data = await response.json() as { analysis: string };

			// Save chat analysis as meeting notes
			const chatNotesPath = await saveMeetingNotes('chat', data.analysis);
			vscode.window.showInformationMessage(`Chat analysis saved to ${chatNotesPath}`);

			// Display the analysis in a new webview
			const panel = vscode.window.createWebviewPanel(
				'washChatAnalysis',
				'Wash Chat Analysis',
				vscode.ViewColumn.One,
				{
					enableScripts: true
				}
			);

			panel.webview.html = getWebviewContent(data.analysis);
		} catch (error) {
			vscode.window.showErrorMessage(`Error analyzing chat: ${error}`);
		}
	});
}

// Function to start continuous chat monitoring
async function startChatMonitoring() {
	if (chatMonitorState.isMonitoring) {
		vscode.window.showInformationMessage('Chat monitoring is already active');
		return;
	}

	// Create or get the continuous chat notes file
	const notesPath = await ensureMeetingNotesFolder();
	const continuousNotesPath = path.join(notesPath, 'continuous-chat-analysis.md');
	chatMonitorState.continuousNotesPath = continuousNotesPath;

	// Initialize or append to the continuous notes file
	if (!fs.existsSync(continuousNotesPath)) {
		const initialContent = `# Continuous Chat Analysis
*Started on ${new Date().toLocaleString()}*

## Conversation Patterns and Insights

`;
		await fs.promises.writeFile(continuousNotesPath, initialContent);
	}

	chatMonitorState.isMonitoring = true;
	chatMonitorState.lastScreenshotTime = Date.now();

	// Start the monitoring interval
	const monitorInterval = setInterval(async () => {
		if (!chatMonitorState.isMonitoring) {
			clearInterval(monitorInterval);
			return;
		}

		try {
			console.log('Capturing screenshot...');
			const screenshot = await captureScreenshot();
			if (!screenshot) {
				console.log('Failed to capture screenshot');
				return;
			}

			console.log('Sending screenshot for analysis...');
			const response = await fetch(`${API_URL}analyze-chat`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ 
					screenshot,
					isContinuous: true,
					lastAnalysisTime: chatMonitorState.lastScreenshotTime
				})
			});

			if (!response.ok) {
				const errorText = await response.text();
				console.error('Server error:', {
					status: response.status,
					statusText: response.statusText,
					error: errorText
				});
				throw new Error(`Server responded with status ${response.status}: ${errorText}`);
			}

			const data = await response.json() as { analysis: string };
			console.log('Received analysis:', data.analysis.substring(0, 100) + '...');
			
			// Append the new analysis to the continuous notes
			const timestamp = new Date().toLocaleString();
			const newContent = `\n### Analysis at ${timestamp}\n\n${data.analysis}\n\n---\n`;
			
			console.log('Appending analysis to continuous notes...');
			await fs.promises.appendFile(continuousNotesPath, newContent);
			chatMonitorState.lastScreenshotTime = Date.now();
			console.log('Analysis appended successfully');

		} catch (error) {
			console.error('Error in continuous chat monitoring:', error);
			vscode.window.showErrorMessage(`Error in chat monitoring: ${error instanceof Error ? error.message : 'Unknown error'}`);
		}
	}, CHAT_MONITOR_INTERVAL);

	vscode.window.showInformationMessage('Started continuous chat monitoring');
}

// Function to stop continuous chat monitoring
function stopChatMonitoring() {
	if (!chatMonitorState.isMonitoring) {
		vscode.window.showInformationMessage('Chat monitoring is not active');
		return;
	}

	chatMonitorState.isMonitoring = false;
	vscode.window.showInformationMessage('Stopped continuous chat monitoring');
}

// Function to generate summary of chat notes
async function generateChatSummary() {
	try {
		const notesPath = await ensureMeetingNotesFolder();
		const continuousNotesPath = path.join(notesPath, 'continuous-chat-analysis.md');

		if (!fs.existsSync(continuousNotesPath)) {
			vscode.window.showErrorMessage('No chat analysis notes found. Start chat monitoring first.');
			return;
		}

		// Read the continuous notes file
		const notesContent = await fs.promises.readFile(continuousNotesPath, 'utf-8');

		// Send the notes to the backend for summarization
		const response = await fetch(`${API_URL}summarize-chat`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({ notes: notesContent })
		});

		if (!response.ok) {
			throw new Error(`Server responded with status ${response.status}`);
		}

		const data = await response.json() as { summary: string };

		// Display the summary in a new webview
		const panel = vscode.window.createWebviewPanel(
			'washChatSummary',
			'Wash Chat Summary',
			vscode.ViewColumn.One,
			{
				enableScripts: true
			}
		);

		panel.webview.html = getWebviewContent(data.summary);
	} catch (error) {
		vscode.window.showErrorMessage(`Error generating chat summary: ${error}`);
	}
}

// This method is called when your extension is deactivated
export function deactivate() {}
