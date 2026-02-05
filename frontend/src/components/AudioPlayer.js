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
    }
  }, [audioFileUrl]);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    const updateTime = () => setCurrentTime(audio.currentTime);
    const updateDuration = () => setDuration(audio.duration);
    const handleLoadStart = () => setIsLoading(true);
    const handleCanPlay = () => setIsLoading(false);
    const handleError = () => {
      setError('Greška pri učitavanju audio fajla');
      setIsLoading(false);
    };

    audio.addEventListener('timeupdate', updateTime);
    audio.addEventListener('loadedmetadata', updateDuration);
    audio.addEventListener('loadstart', handleLoadStart);
    audio.addEventListener('canplay', handleCanPlay);
    audio.addEventListener('error', handleError);

    return () => {
      audio.removeEventListener('timeupdate', updateTime);
      audio.removeEventListener('loadedmetadata', updateDuration);
      audio.removeEventListener('loadstart', handleLoadStart);
      audio.removeEventListener('canplay', handleCanPlay);
      audio.removeEventListener('error', handleError);
    };
  }, []);

  const togglePlayPause = () => {
    const audio = audioRef.current;
    if (!audio) return;

    if (isPlaying) {
      audio.pause();
    } else {
      audio.play().catch(err => {
        console.error('Audio playback error:', err);
        setError('Greška pri reprodukciji: ' + err.message);
      });
    }
    setIsPlaying(!isPlaying);
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
    // If song has direct audio file URL, use it
    if (audioFileUrl) {
      return audioFileUrl;
    }
    
    // Otherwise, use the streaming endpoint
    return api.getStreamUrl(songId);
  };

  return (
    <div className="audio-player">
      <div className="audio-player-header">
        <h4>{songName || 'Audio Player'}</h4>
      </div>
      
      {error && <div className="audio-error">{error}</div>}
      
      <audio
        ref={audioRef}
        src={getAudioUrl()}
        preload="metadata"
        controls
        style={{ width: '100%', marginBottom: '10px' }}
      />
      
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
