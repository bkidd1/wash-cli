import express from 'express';
import cors from 'cors';
import dotenv from 'dotenv';
import OpenAI from 'openai';
import jwt from 'jsonwebtoken';

dotenv.config();

const app = express();
const port = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json());

// JWT secret for authentication
const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key';

// Initialize OpenAI client
const openai = new OpenAI({
    apiKey: process.env.OPENAI_API_KEY
});

// Middleware to verify JWT token
const authenticateToken = (req: express.Request, res: express.Response, next: express.NextFunction) => {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1];

    if (!token) {
        return res.status(401).json({ error: 'No token provided' });
    }

    jwt.verify(token, JWT_SECRET, (err: any) => {
        if (err) {
            return res.status(403).json({ error: 'Invalid token' });
        }
        next();
    });
};

// Endpoint to analyze code
app.post('/analyze', authenticateToken, async (req, res) => {
    try {
        const { code } = req.body;

        if (!code) {
            return res.status(400).json({ error: 'No code provided' });
        }

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

        res.json({ analysis: response.choices[0].message.content });
    } catch (error) {
        console.error('Error analyzing code:', error);
        res.status(500).json({ error: 'Failed to analyze code' });
    }
});

// Endpoint to generate a JWT token
app.post('/auth', (req, res) => {
    const { clientId } = req.body;
    
    if (!clientId) {
        return res.status(400).json({ error: 'Client ID required' });
    }

    const token = jwt.sign({ clientId }, JWT_SECRET, { expiresIn: '24h' });
    res.json({ token });
});

app.listen(port, () => {
    console.log(`Server running on port ${port}`);
}); 