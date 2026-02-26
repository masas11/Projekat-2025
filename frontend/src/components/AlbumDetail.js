import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api from '../services/api';

const AlbumDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [album, setAlbum] = useState(null);
  const [songs, setSongs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadAlbum();
    loadSongs();
  }, [id]);

  const loadAlbum = async () => {
    try {
      const data = await api.getAlbum(id);
      setAlbum(data);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju albuma');
    } finally {
      setLoading(false);
    }
  };

  const loadSongs = async () => {
    try {
      const data = await api.getSongsByAlbum(id);
      setSongs(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error loading songs:', err);
    }
  };

  const formatDuration = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (loading) {
    return (
      <div className="container" style={{ paddingTop: '60px', textAlign: 'center' }}>
        <div style={{ fontSize: '48px', marginBottom: '20px', animation: 'spin 1s linear infinite' }}>⏳</div>
        <p style={{ fontSize: '18px', color: '#666' }}>Učitavanje albuma...</p>
      </div>
    );
  }

  if (error || !album) {
    return (
      <div style={{ minHeight: 'calc(100vh - 80px)', paddingTop: '40px', paddingBottom: '40px' }}>
        <div className="container">
          <div style={{
            background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
            backdropFilter: 'blur(10px)',
            borderRadius: '20px',
            padding: '40px',
            boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
            border: '1px solid rgba(255,255,255,0.3)',
            textAlign: 'center'
          }}>
            <div className="error" style={{ marginBottom: '20px' }}>
              <span>⚠️</span>
              <span>{error || 'Album nije pronađen'}</span>
            </div>
            <button className="btn btn-secondary" onClick={() => navigate('/albums')}>
              ← Nazad na albume
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div style={{ minHeight: 'calc(100vh - 80px)', paddingTop: '40px', paddingBottom: '40px' }}>
      <div className="container">
        {/* Album Header */}
        <div style={{
          background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
          backdropFilter: 'blur(10px)',
          borderRadius: '20px',
          padding: '40px',
          marginBottom: '30px',
          boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
          border: '1px solid rgba(255,255,255,0.3)'
        }}>
          <button 
            className="btn btn-secondary" 
            onClick={() => navigate('/albums')} 
            style={{ 
              marginBottom: '30px',
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '12px 20px'
            }}
          >
            ← Nazad na albume
          </button>

          <div style={{ display: 'flex', alignItems: 'flex-start', gap: '30px', flexWrap: 'wrap' }}>
            {/* Album Cover */}
            <div style={{
              width: '200px',
              height: '200px',
              borderRadius: '20px',
              background: `linear-gradient(135deg, 
                hsl(${(album.id.charCodeAt(0) * 137.508) % 360}, 70%, 60%) 0%, 
                hsl(${(album.id.charCodeAt(0) * 137.508 + 60) % 360}, 70%, 50%) 100%)`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '80px',
              boxShadow: '0 10px 30px rgba(0,0,0,0.2)',
              flexShrink: 0
            }}>
              💿
            </div>

            {/* Album Info */}
            <div style={{ flex: 1, minWidth: '300px' }}>
              <h1 style={{
                margin: '0 0 16px 0',
                fontSize: '42px',
                fontWeight: '700',
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text'
              }}>
                {album.name}
              </h1>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                {album.genre && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span style={{ fontSize: '18px' }}>🎵</span>
                    <span 
                      className="genre-tag"
                      style={{
                        padding: '8px 16px',
                        background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
                        color: '#667eea',
                        border: '1px solid rgba(102, 126, 234, 0.2)',
                        fontWeight: '600',
                        fontSize: '14px',
                        borderRadius: '16px'
                      }}
                    >
                      {album.genre}
                    </span>
                  </div>
                )}

                {album.releaseDate && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span style={{ fontSize: '18px' }}>📅</span>
                    <span style={{ fontSize: '16px', color: '#666', fontWeight: '500' }}>Datum izdavanja:</span>
                    <span style={{ fontSize: '16px', fontWeight: '600', color: '#333' }}>
                      {new Date(album.releaseDate).toLocaleDateString('sr-RS', {
                        year: 'numeric',
                        month: 'long',
                        day: 'numeric'
                      })}
                    </span>
                  </div>
                )}

                {songs.length > 0 && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginTop: '8px' }}>
                    <span style={{ fontSize: '18px' }}>🎵</span>
                    <span style={{ fontSize: '16px', color: '#666', fontWeight: '500' }}>
                      {songs.length} {songs.length === 1 ? 'pesma' : songs.length < 5 ? 'pesme' : 'pesama'}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Songs List */}
        {songs.length > 0 ? (
          <div style={{
            background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
            backdropFilter: 'blur(10px)',
            borderRadius: '20px',
            padding: '40px',
            boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
            border: '1px solid rgba(255,255,255,0.3)'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '30px' }}>
              <span style={{ fontSize: '32px' }}>🎵</span>
              <h2 style={{
                margin: 0,
                fontSize: '28px',
                fontWeight: '700',
                color: '#333'
              }}>
                Pesme
              </h2>
              <span style={{
                marginLeft: 'auto',
                padding: '6px 14px',
                background: 'rgba(102, 126, 234, 0.1)',
                borderRadius: '12px',
                fontSize: '14px',
                fontWeight: '600',
                color: '#667eea'
              }}>
                {songs.length} {songs.length === 1 ? 'pesma' : songs.length < 5 ? 'pesme' : 'pesama'}
              </span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              {songs.map((song, index) => (
                <div
                  key={song.id}
                  style={{
                    padding: '20px',
                    background: 'rgba(102, 126, 234, 0.03)',
                    borderRadius: '16px',
                    border: '2px solid transparent',
                    cursor: 'pointer',
                    transition: 'all 0.3s ease',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '20px'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = 'rgba(102, 126, 234, 0.08)';
                    e.currentTarget.style.borderColor = 'rgba(102, 126, 234, 0.3)';
                    e.currentTarget.style.transform = 'translateX(8px)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = 'rgba(102, 126, 234, 0.03)';
                    e.currentTarget.style.borderColor = 'transparent';
                    e.currentTarget.style.transform = 'translateX(0)';
                  }}
                  onClick={() => navigate(`/songs/${song.id}`)}
                >
                  {/* Track Number */}
                  <div style={{
                    width: '40px',
                    height: '40px',
                    borderRadius: '10px',
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: 'white',
                    fontWeight: '700',
                    fontSize: '16px',
                    flexShrink: 0
                  }}>
                    {index + 1}
                  </div>

                  {/* Song Info */}
                  <div style={{ flex: 1 }}>
                    <h4 style={{
                      margin: '0 0 8px 0',
                      fontSize: '20px',
                      fontWeight: '600',
                      color: '#333'
                    }}>
                      {song.name}
                    </h4>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '16px', flexWrap: 'wrap' }}>
                      {song.duration && (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <span style={{ fontSize: '14px' }}>⏱️</span>
                          <span style={{ fontSize: '14px', color: '#666' }}>
                            {formatDuration(song.duration)}
                          </span>
                        </div>
                      )}
                      {song.genre && (
                        <span 
                          className="genre-tag"
                          style={{
                            padding: '4px 12px',
                            background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
                            color: '#667eea',
                            border: '1px solid rgba(102, 126, 234, 0.2)',
                            fontWeight: '500',
                            fontSize: '12px',
                            borderRadius: '12px'
                          }}
                        >
                          {song.genre}
                        </span>
                      )}
                      {/* API Composition: Average Rating and Rating Count */}
                      {(song.averageRating !== undefined || song.ratingCount !== undefined) && (
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <span style={{ fontSize: '14px' }}>⭐</span>
                          {song.averageRating !== undefined && song.averageRating > 0 ? (
                            <span style={{ fontSize: '14px', color: '#666' }}>
                              <strong style={{ color: '#667eea' }}>{song.averageRating.toFixed(1)}</strong>
                              {song.ratingCount !== undefined && song.ratingCount > 0 && (
                                <span style={{ marginLeft: '4px', fontSize: '12px', color: '#999' }}>
                                  ({song.ratingCount})
                                </span>
                              )}
                            </span>
                          ) : (
                            <span style={{ fontSize: '12px', color: '#999', fontStyle: 'italic' }}>
                              Nema ocena
                            </span>
                          )}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Play Icon */}
                  <div style={{
                    fontSize: '24px',
                    opacity: 0.4,
                    transition: 'all 0.3s ease'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.opacity = 1;
                    e.currentTarget.style.transform = 'scale(1.2)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.opacity = 0.4;
                    e.currentTarget.style.transform = 'scale(1)';
                  }}
                  >
                    ▶️
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div style={{
            background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
            backdropFilter: 'blur(10px)',
            borderRadius: '20px',
            padding: '60px 40px',
            textAlign: 'center',
            boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
            border: '1px solid rgba(255,255,255,0.3)'
          }}>
            <div style={{ fontSize: '64px', marginBottom: '20px' }}>🎵</div>
            <h3 style={{ fontSize: '24px', fontWeight: '600', color: '#333', marginBottom: '12px' }}>
              Nema pesama u ovom albumu
            </h3>
            <p style={{ color: '#666', fontSize: '16px' }}>
              Pesme će biti dodate uskoro!
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default AlbumDetail;
