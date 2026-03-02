import React, { useState, useRef, useEffect } from 'react';
import api from '../services/api';
import './AudioPlayer.css';

const AudioPlayer = ({ songId, songName, audioFileUrl }) => {
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [cacheBuster, setCacheBuster] = useState(Date.now());
  
  const audioRef = useRef(null);
  const progressBarRef = useRef(null);

  // Validate URL on mount and when audioFileUrl changes
  useEffect(() => {
    if (audioFileUrl) {
      if (audioFileUrl.includes('youtube.com') || audioFileUrl.includes('youtu.be')) {
        setError('YouTube URL-ovi nisu podržani. Koristite direktn link ka MP3 fajlu.');
      } else {
        setError(''); // Clear error if URL is valid
      }
      // Update cache buster when audioFileUrl changes
      setCacheBuster(Date.now());
    }
  }, [audioFileUrl]);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    const updateTime = () => setCurrentTime(audio.currentTime);
    const updateDuration = () => setDuration(audio.duration);
    const handleLoadStart = () => setIsLoading(true);
    const handleCanPlay = () => setIsLoading(false);
    const handleError = (e) => {
      const audio = audioRef.current;
      if (audio && audio.error) {
        let errorMsg = 'Greška pri učitavanju audio fajla';
        switch (audio.error.code) {
          case audio.error.MEDIA_ERR_ABORTED:
            errorMsg = 'Reprodukcija je prekinuta';
            break;
          case audio.error.MEDIA_ERR_NETWORK:
            errorMsg = 'Greška u mreži - audio fajl nije dostupan';
            break;
          case audio.error.MEDIA_ERR_DECODE:
            errorMsg = 'Greška pri dekodiranju audio fajla';
            break;
          case audio.error.MEDIA_ERR_SRC_NOT_SUPPORTED:
            errorMsg = 'Audio fajl nije dostupan ili format nije podržan';
            break;
        }
        setError(errorMsg);
      } else {
        setError('Greška pri učitavanju audio fajla');
      }
      setIsLoading(false);
    };

    audio.addEventListener('timeupdate', updateTime);
    audio.addEventListener('loadedmetadata', updateDuration);
    audio.addEventListener('loadstart', handleLoadStart);
    audio.addEventListener('canplay', handleCanPlay);
    audio.addEventListener('error', (e) => handleError(e));

    return () => {
      audio.removeEventListener('timeupdate', updateTime);
      audio.removeEventListener('loadedmetadata', updateDuration);
      audio.removeEventListener('loadstart', handleLoadStart);
      audio.removeEventListener('canplay', handleCanPlay);
      audio.removeEventListener('error', handleError);
    };
  }, []);

  const togglePlayPause = async () => {
    const audio = audioRef.current;
    if (!audio) return;

    if (isPlaying) {
      audio.pause();
      setIsPlaying(false);
    } else {
      // Check if audio source is valid before trying to play
      if (!audio.src || audio.src === '') {
        setError('Audio fajl nije dostupan za ovu pesmu');
        return;
      }
      
      try {
        await audio.play();
        setIsPlaying(true);
      } catch (err) {
        console.error('Audio playback error:', err);
        if (err.name === 'NotSupportedError' || err.message.includes('no supported sources')) {
          setError('Audio fajl nije dostupan. Molimo upload-ujte audio fajl za ovu pesmu.');
        } else {
          setError('Greška pri reprodukciji: ' + err.message);
        }
        setIsPlaying(false);
      }
    }
  };

  const handleSeek = (e) => {
    const audio = audioRef.current;
    const progressBar = progressBarRef.current;
    if (!audio || !progressBar) return;

    const rect = progressBar.getBoundingClientRect();
    const percent = (e.clientX - rect.left) / rect.width;
    const newTime = percent * duration;
    
    audio.currentTime = newTime;
    setCurrentTime(newTime);
  };

  const handleVolumeChange = (e) => {
    const audio = audioRef.current;
    const newVolume = parseFloat(e.target.value);
    
    if (audio) {
      audio.volume = newVolume;
      setVolume(newVolume);
    }
  };

  const formatTime = (seconds) => {
    if (isNaN(seconds)) return '0:00';
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const getAudioUrl = () => {
    if (!songId) {
      console.warn('No songId provided to AudioPlayer');
      return null;
    }
    
    // If song has HDFS path (starts with /audio/), use stream endpoint with cache-busting
    if (audioFileUrl && (audioFileUrl.startsWith('/audio/') || audioFileUrl.startsWith('hdfs://'))) {
      const baseUrl = api.getStreamUrl(songId);
      // Remove existing timestamp and add new one
      const url = baseUrl.split('&t=')[0].split('?t=')[0];
      const separator = url.includes('?') ? '&' : '?';
      return `${url}${separator}t=${cacheBuster}`;
    }
    
    // If song has external URL (http/https), use it directly with cache-busting
    if (audioFileUrl && (audioFileUrl.startsWith('http://') || audioFileUrl.startsWith('https://'))) {
      // Add cache-busting parameter to external URLs too
      const separator = audioFileUrl.includes('?') ? '&' : '?';
      return `${audioFileUrl}${separator}t=${cacheBuster}`;
    }
    
    // Otherwise, use the streaming endpoint with cache-busting
    const baseUrl = api.getStreamUrl(songId);
    const url = baseUrl.split('&t=')[0].split('?t=')[0];
    const separator = url.includes('?') ? '&' : '?';
    return `${url}${separator}t=${cacheBuster}`;
  };

  return (
    <div className="audio-player">
      <div className="audio-player-header">
        <h4>{songName || 'Audio Player'}</h4>
      </div>
      
      {error && <div className="audio-error">{error}</div>}
      
      {getAudioUrl() ? (
        <audio
          key={audioFileUrl || songId} // Force re-render when audioFileUrl changes
          ref={audioRef}
          src={getAudioUrl()}
          preload="metadata"
          controls
          style={{ width: '100%', marginBottom: '10px' }}
        />
      ) : (
        <div className="audio-error">Audio URL nije dostupan</div>
      )}
      
      <div className="audio-controls">
        <button 
          className="play-pause-btn"
          onClick={togglePlayPause}
          disabled={isLoading}
        >
          {isLoading ? (
            <span className="loading-spinner">⏳</span>
          ) : isPlaying ? (
            '⏸️'
          ) : (
            '▶️'
          )}
        </button>
        
        <div className="time-display">
          <span>{formatTime(currentTime)}</span>
        </div>
        
        <div 
          className="progress-bar"
          ref={progressBarRef}
          onClick={handleSeek}
        >
          <div 
            className="progress-fill"
            style={{ width: `${(currentTime / duration) * 100 || 0}%` }}
          />
        </div>
        
        <div className="time-display">
          <span>{formatTime(duration)}</span>
        </div>
        
        <div className="volume-control">
          <span>🔊</span>
          <input
            type="range"
            min="0"
            max="1"
            step="0.1"
            value={volume}
            onChange={handleVolumeChange}
            className="volume-slider"
          />
        </div>
      </div>
    </div>
  );
};

export default AudioPlayer;
