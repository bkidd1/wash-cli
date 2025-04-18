import fetch from 'node-fetch';

const API_URL = 'http://localhost:3000/';

// JWT token using the same secret as the server
const JWT_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkxvY2FsIERldiIsImlhdCI6MTUxNjIzOTAyMn0.your_jwt_secret_here';

async function testAPI() {
    try {
        console.log('\n=== Local API Endpoint Test ===');
        console.log('Testing URL:', `${API_URL}analyze`);
        
        // Test 1: Check if endpoint is reachable
        console.log('\n1. Testing endpoint reachability...');
        const startTime = Date.now();
        const response = await fetch(`${API_URL}analyze`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
            },
            body: JSON.stringify({ 
                code: 'console.log("Hello, World!");' 
            })
        });
        const endTime = Date.now();
        console.log(`Response time: ${endTime - startTime}ms`);
        console.log('Response status:', response.status);
        
        // Test 2: Check CORS headers
        console.log('\n2. Checking CORS headers...');
        const corsHeaders = {
            'Access-Control-Allow-Origin': response.headers.get('Access-Control-Allow-Origin'),
            'Access-Control-Allow-Methods': response.headers.get('Access-Control-Allow-Methods'),
            'Access-Control-Allow-Headers': response.headers.get('Access-Control-Allow-Headers'),
        };
        console.log('CORS Headers:', JSON.stringify(corsHeaders, null, 2));
        
        // Test 3: Check all response headers
        console.log('\n3. All Response Headers:');
        const headers: Record<string, string> = {};
        response.headers.forEach((value, key) => {
            headers[key] = value;
        });
        console.log(JSON.stringify(headers, null, 2));
        
        // Test 4: Check response body
        console.log('\n4. Response Body:');
        const text = await response.text();
        try {
            const json = JSON.parse(text);
            console.log(JSON.stringify(json, null, 2));
        } catch (e) {
            console.log('Raw response:', text);
        }

        // Test 5: Check preflight request
        console.log('\n5. Testing OPTIONS request...');
        const optionsResponse = await fetch(`${API_URL}analyze`, {
            method: 'OPTIONS',
            headers: {
                'Origin': 'http://localhost:3000',
                'Access-Control-Request-Method': 'POST',
                'Access-Control-Request-Headers': 'Content-Type,Authorization',
            }
        });
        console.log('OPTIONS Response Status:', optionsResponse.status);
        console.log('OPTIONS Headers:', JSON.stringify(Object.fromEntries(optionsResponse.headers.entries()), null, 2));

    } catch (error) {
        console.error('\n=== Error Details ===');
        console.error('Error testing API:', error);
        if (error instanceof Error) {
            console.error('Error message:', error.message);
            console.error('Error stack:', error.stack);
            
            if (error.name === 'TypeError' && error.message.includes('fetch failed')) {
                console.error('\nNetwork Error Details:');
                console.error('- Make sure the local server is running');
                console.error('- Check if the server is listening on port 3000');
                console.error('- Verify the server is accepting connections');
            }
        }
    }
}

// Run the tests
console.log('Starting local API tests...');
testAPI().then(() => {
    console.log('\n=== Test Complete ===');
}).catch(error => {
    console.error('\n=== Test Failed ===');
    console.error(error);
}); 