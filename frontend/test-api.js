const API_BASE = 'http://localhost:8083';

async function doApiRequest(path) {
    const response = await fetch(`${API_BASE}${path}`);
    const json = await response.json();
    
    if (!json.ok) {
        throw new Error(json.error || 'Unknown error');
    }
    
    return json.result;
}

async function testTraceAPI() {
    console.log('测试 Trace API...');
    try {
        const result = await doApiRequest('/api/v1/trace/ethereum/0x1234567890123456789012345678901234567890123456789012345678901234');
        console.log('✅ Trace API 测试成功');
        console.log('返回数据结构:', Object.keys(result));
        console.log('Chain:', result.chain);
        console.log('Txhash:', result.txhash);
        console.log('Entrypoint type:', result.entrypoint.type);
        console.log('Addresses count:', Object.keys(result.addresses).length);
    } catch (err) {
        console.error('❌ Trace API 测试失败:', err.message);
    }
}

async function testStorageAPI() {
    console.log('\n测试 Storage API...');
    try {
        const result = await doApiRequest('/api/v1/storage/ethereum/0x1234567890123456789012345678901234567890/0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd');
        console.log('✅ Storage API 测试成功');
        console.log('返回数据结构:', Object.keys(result));
        console.log('AllStructs count:', result.allStructs.length);
        console.log('Arrays count:', result.arrays.length);
        console.log('Structs count:', result.structs.length);
        console.log('Slots count:', Object.keys(result.slots).length);
    } catch (err) {
        console.error('❌ Storage API 测试失败:', err.message);
    }
}

async function main() {
    console.log('🚀 开始测试前端 API 调用...\n');
    
    await testTraceAPI();
    await testStorageAPI();
    
    console.log('\n✨ 测试完成！');
}

main().catch(console.error); 