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
            console.log('Testing trace API...');
            const result = await doApiRequest('/api/v1/trace/avalanche/0xb778319c7afd1322f944e41e48cdaad79c83f3385b5932e869f1499f61ebc114');
            console.log('Trace API result:', result);
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
                <div style={{ 
                    padding: '10px', 
                    backgroundColor: '#d4edda', 
                    color: '#155724', 
                    border: '1px solid #c3e6cb',
                    borderRadius: '5px',
                    marginBottom: '20px'
                }}>
                    <h3>Trace API 结果:</h3>
                    <pre style={{ whiteSpace: 'pre-wrap', fontSize: '12px' }}>
                        {JSON.stringify(traceResult, null, 2)}
                    </pre>
                </div>
            )}

            {storageResult && (
                <div style={{ 
                    padding: '10px', 
                    backgroundColor: '#d1ecf1', 
                    color: '#0c5460', 
                    border: '1px solid #bee5eb',
                    borderRadius: '5px',
                    marginBottom: '20px'
                }}>
                    <h3>Storage API 结果:</h3>
                    <pre style={{ whiteSpace: 'pre-wrap', fontSize: '12px' }}>
                        {JSON.stringify(storageResult, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    );
} 