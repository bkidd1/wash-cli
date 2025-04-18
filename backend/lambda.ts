import { APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';
import OpenAI from 'openai';
import { SecretsManager } from 'aws-sdk';

const secretsManager = new SecretsManager();

// Function to get the OpenAI API key from Secrets Manager
async function getOpenAIKey(): Promise<string> {
    try {
        const secret = await secretsManager.getSecretValue({
            SecretId: 'wash-openai-api-key'
        }).promise();
        
        if (secret.SecretString) {
            const secretObject = JSON.parse(secret.SecretString);
            return secretObject.OPENAI_API_KEY;
        }
        throw new Error('No secret string found');
    } catch (error) {
        console.error('Error fetching OpenAI API key:', error);
        throw error;
    }
}

// Initialize OpenAI client with a function to get the API key
const openai = new OpenAI({
    apiKey: process.env.OPENAI_API_KEY
});

export const handler = async (event: APIGatewayProxyEvent): Promise<APIGatewayProxyResult> => {
    // Handle CORS
    const headers = {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Credentials': true,
        'Content-Type': 'application/json'
    };

    // Handle preflight requests
    if (event.httpMethod === 'OPTIONS') {
        return {
            statusCode: 200,
            headers: {
                ...headers,
                'Access-Control-Allow-Methods': 'POST,OPTIONS',
                'Access-Control-Allow-Headers': 'Content-Type'
            },
            body: ''
        };
    }

    try {
        // Get the OpenAI API key
        const apiKey = await getOpenAIKey();
        openai.apiKey = apiKey;

        // Parse the request body
        const body = JSON.parse(event.body || '{}');
        const { code } = body;

        if (!code) {
            return {
                statusCode: 400,
                headers,
                body: JSON.stringify({ error: 'No code provided' })
            };
        }

        // Call OpenAI API
        const response = await openai.chat.completions.create({
            model: "gpt-4-vision-preview",
            messages: [
                {
                    role: "system",
                    content: "You are an expert coding assistant. Analyze the provided code and suggest alternative approaches or pathways the developer could have taken. Consider different architectural patterns, libraries, or methodologies that might be more efficient or maintainable."
                },
                {
                    role: "user",
                    content: [
                        {
                            type: "text",
                            text: "Here's the current code implementation. Please suggest alternative approaches or pathways:"
                        },
                        {
                            type: "text",
                            text: code
                        }
                    ]
                }
            ],
            max_tokens: 1000
        });

        return {
            statusCode: 200,
            headers,
            body: JSON.stringify({ 
                analysis: response.choices[0].message.content 
            })
        };
    } catch (error) {
        console.error('Error:', error);
        return {
            statusCode: 500,
            headers,
            body: JSON.stringify({ error: 'Failed to analyze code' })
        };
    }
}; 