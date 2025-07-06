import { useState } from 'react';
import { doApiRequest } from '../components/tracer/api';

export default function TestAPI() {
    const [traceResult, setTraceResult] = useState(null);
    const [storageResult, setStorageResult] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const testTraceAPI = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await doApiRequest('/api/v1/trace/ethereum/0x1234567890123456789012345678901234567890123456789012345678901234');
            setTraceResult(result);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const testStorageAPI = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await doApiRequest('/api/v1/storage/ethereum/0x1234567890123456789012345678901234567890/0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd');
            setStorageResult(result);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
            <h1>API 测试页面</h1>
            <p>测试前端是否能正确调用本地 tx-tracer-srv API</p>
            
            <div style={{ marginBottom: '20px' }}>
                <button 
                    onClick={testTraceAPI}
                    disabled={loading}
                    style={{ 
                        padding: '10px 20px', 
                        marginRight: '10px',
                        backgroundColor: '#007bff',
                        color: 'white',
                        border: 'none',
                        borderRadius: '5px',
                        cursor: loading ? 'not-allowed' : 'pointer'
                    }}
                >
                    {loading ? '测试中...' : '测试 Trace API'}
                </button>
                
                <button 
                    onClick={testStorageAPI}
                    disabled={loading}
                    style={{ 
                        padding: '10px 20px',
                        backgroundColor: '#28a745',
                        color: 'white',
                        border: 'none',
                        borderRadius: '5px',
                        cursor: loading ? 'not-allowed' : 'pointer'
                    }}
                >
                    {loading ? '测试中...' : '测试 Storage API'}
                </button>
            </div>

            {error && (
                <div style={{ 
                    padding: '10px', 
                    backgroundColor: '#f8d7da', 
                    color: '#721c24', 
                    border: '1px solid #f5c6cb',
                    borderRadius: '5px',
                    marginBottom: '20px'
                }}>
                    <strong>错误:</strong> {error}
                </div>
            )}

            {traceResult && (
                <div style={{ marginBottom: '20px' }}>
                    <h3>Trace API 结果:</h3>
                    <pre style={{ 
                        backgroundColor: '#f8f9fa', 
                        padding: '10px', 
                        borderRadius: '5px',
                        overflow: 'auto',
                        maxHeight: '300px'
                    }}>
                        {JSON.stringify(traceResult, null, 2)}
                    </pre>
                </div>
            )}

            {storageResult && (
                <div>
                    <h3>Storage API 结果:</h3>
                    <pre style={{ 
                        backgroundColor: '#f8f9fa', 
                        padding: '10px', 
                        borderRadius: '5px',
                        overflow: 'auto',
                        maxHeight: '300px'
                    }}>
                        {JSON.stringify(storageResult, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    );
} 