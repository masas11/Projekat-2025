import React, { useState } from 'react';

const UrlTester = () => {
  const [url, setUrl] = useState('');
  const [testResult, setTestResult] = useState('');

  const testUrl = () => {
    if (!url) {
      setTestResult('Unesi URL prvo');
      return;
    }

    const audio = new Audio();
    
    audio.addEventListener('canplaythrough', () => {
      setTestResult('✅ URL je validan i može se reprodukovati');
    });
    
    audio.addEventListener('error', () => {
      setTestResult('❌ URL nije validan ili ne može se reprodukovati');
    });

    audio.src = url;
  };

  const sampleUrls = [
    'https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3',
    'https://file-examples.com/storage/fe86ead4707ced6b9ef4c06/content/2017/11/file_example_MP3_700KB.mp3',
    'https://www.soundhelix.com/examples/mp3/SoundHelix-Song-2.mp3',
  ];

  return (
    <div style={{ padding: '20px', maxWidth: '600px', margin: '0 auto' }}>
      <h3>Audio URL Tester</h3>
      
      <div style={{ marginBottom: '20px' }}>
        <label>Testiraj audio URL:</label>
        <input
          type="text"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com/song.mp3"
          style={{ 
            width: '100%', 
            padding: '8px', 
            margin: '10px 0',
            border: '1px solid #ddd',
            borderRadius: '4px'
          }}
        />
        <button onClick={testUrl} className="btn btn-primary">
          Testiraj URL
        </button>
      </div>

      {testResult && (
        <div style={{ 
          padding: '10px', 
          margin: '10px 0',
          backgroundColor: testResult.includes('✅') ? '#d4edda' : '#f8d7da',
          borderRadius: '4px'
        }}>
          {testResult}
        </div>
      )}

      <div style={{ marginTop: '30px' }}>
        <h4>Sample URL-ovi za testiranje:</h4>
        {sampleUrls.map((sampleUrl, index) => (
          <div key={index} style={{ margin: '10px 0' }}>
            <button 
              onClick={() => setUrl(sampleUrl)}
              className="btn btn-secondary"
              style={{ fontSize: '0.8em' }}
            >
              Koristi Sample {index + 1}
            </button>
            <div style={{ fontSize: '0.8em', color: '#666', marginTop: '5px' }}>
              {sampleUrl}
            </div>
          </div>
        ))}
      </div>

      <div style={{ marginTop: '30px', padding: '15px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
        <h4>Kako naći dobre URL-ove:</h4>
        <ul>
          <li>Traži "free mp3 download" ili "royalty free music"</li>
          <li>Koristi Creative Commons sajtove</li>
          <li>Izbegavaj YouTube URL-ove (ne rade)</li>
          <li>Proveri da URL završava sa .mp3, .wav, .ogg</li>
          <li>Koristi ovaj tester pre dodavanja u pesmu</li>
        </ul>
      </div>
    </div>
  );
};

export default UrlTester;
