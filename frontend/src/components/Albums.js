import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Albums = () => {
  const [albums, setAlbums] = useState([]);
  const [filteredAlbums, setFilteredAlbums] = useState([]);
  const [artists, setArtists] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [editingAlbum, setEditingAlbum] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedGenre, setSelectedGenre] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    releaseDate: '',
    genre: '',
    selectedArtistIds: [],
  });
  const { isAdmin, user, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [subscriptionMessage, setSubscriptionMessage] = useState('');
  const [isSubscribing, setIsSubscribing] = useState(false);
  const [subscribedGenres, setSubscribedGenres] = useState([]);

  useEffect(() => {
    loadAlbums();
    loadArtists();
    if (isAuthenticated && user) {
      loadSubscriptions();
    }
  }, [isAuthenticated, user]);

  useEffect(() => {
    filterAlbums();
  }, [albums, searchTerm, selectedGenre]);

  const filterAlbums = () => {
    let filtered = albums;

    // Filter by search term
    if (searchTerm) {
      filtered = filtered.filter(album =>
        album.name.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Filter by genre
    if (selectedGenre) {
      filtered = filtered.filter(album =>
        album.genre === selectedGenre
      );
    }

    setFilteredAlbums(filtered);
  };

  const loadAlbums = async () => {
    try {
      const data = await api.getAlbums();
      setAlbums(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju albuma');
    } finally {
      setLoading(false);
    }
  };

  const loadArtists = async () => {
    try {
      const data = await api.getArtists();
      setArtists(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error loading artists:', err);
    }
  };

  const loadSubscriptions = async () => {
    if (!isAuthenticated || !user) return;

    try {
      const subscriptions = await api.getSubscriptions();
      if (Array.isArray(subscriptions)) {
        const genres = subscriptions
          .filter(sub => sub && sub.type === 'genre')
          .map(sub => sub.genre);
        setSubscribedGenres(genres);
      } else {
        setSubscribedGenres([]);
      }
    } catch (err) {
      console.error('Error loading subscriptions:', err);
      setSubscribedGenres([]);
    }
  };

  const handleSubscribeToGenre = async (genre) => {
    if (!isAuthenticated || !user) {
      setError('Morate biti prijavljeni da biste se pretplatili na žanr');
      return;
    }

    if (!genre) {
      setError('Izaberite žanr za pretplatu');
      return;
    }

    const isSubscribed = subscribedGenres.includes(genre);

    setIsSubscribing(true);
    setSubscriptionMessage('');
    setError('');

    try {
      if (isSubscribed) {
        await api.unsubscribeFromGenre(genre, user.id);
        setSubscribedGenres(subscribedGenres.filter(g => g !== genre));
        setSubscriptionMessage(`Uspešno ste se odjavili sa pretplate na žanr: ${genre}!`);
      } else {
        await api.subscribeToGenre(genre, user.id);
        setSubscribedGenres([...subscribedGenres, genre]);
        setSubscriptionMessage(`Uspešno ste se pretplatili na žanr: ${genre}!`);
      }
      setTimeout(() => setSubscriptionMessage(''), 3000);
    } catch (err) {
      if (err.message && err.message.includes('Already subscribed')) {
        setSubscribedGenres([...subscribedGenres, genre]);
        setError('Već ste pretplaćeni na ovaj žanr');
      } else {
        setError(err.message || 'Greška pri pretplati na žanr');
      }
    } finally {
      setIsSubscribing(false);
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };

  const handleArtistSelect = (e) => {
    const selectedOptions = Array.from(e.target.selectedOptions, option => option.value);
    setFormData({
      ...formData,
      selectedArtistIds: selectedOptions,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    const albumData = {
      name: formData.name,
      releaseDate: formData.releaseDate ? new Date(formData.releaseDate).toISOString() : new Date().toISOString(),
      genre: formData.genre,
      artistIds: formData.selectedArtistIds,
    };

    try {
      if (editingAlbum) {
        await api.updateAlbum(editingAlbum.id, albumData);
      } else {
        await api.createAlbum(albumData);
      }
      setShowForm(false);
      setEditingAlbum(null);
      setFormData({ name: '', releaseDate: '', genre: '', selectedArtistIds: [] });
      loadAlbums();
    } catch (err) {
      setError(err.message || 'Greška pri čuvanju albuma');
    }
  };

  const handleEdit = (album) => {
    setEditingAlbum(album);
    setFormData({
      name: album.name,
      releaseDate: album.releaseDate ? new Date(album.releaseDate).toISOString().split('T')[0] : '',
      genre: album.genre || '',
      selectedArtistIds: album.artistIds || album.artistIDs || [],
    });
    setShowForm(true);
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingAlbum(null);
    setFormData({ name: '', releaseDate: '', genre: '', selectedArtistIds: [] });
  };

  // Get all unique genres from albums
  const allGenres = [...new Set(albums.map(album => album.genre).filter(Boolean))].sort();

  if (loading) {
    return (
      <div className="container" style={{ paddingTop: '60px', textAlign: 'center' }}>
        <div style={{ fontSize: '48px', marginBottom: '20px', animation: 'spin 1s linear infinite' }}>⏳</div>
        <p style={{ fontSize: '18px', color: '#666' }}>Učitavanje albuma...</p>
      </div>
    );
  }

  return (
    <div style={{ minHeight: 'calc(100vh - 80px)', paddingTop: '40px', paddingBottom: '40px' }}>
      <div className="container">
        {/* Header Section */}
        <div style={{
          background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
          backdropFilter: 'blur(10px)',
          borderRadius: '20px',
          padding: '40px',
          marginBottom: '30px',
          boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
          border: '1px solid rgba(255,255,255,0.3)'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '20px' }}>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '8px' }}>
                <div style={{
                  width: '60px',
                  height: '60px',
                  borderRadius: '16px',
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '32px',
                  boxShadow: '0 8px 20px rgba(102, 126, 234, 0.3)'
                }}>
                  💿
                </div>
                <div>
                  <h1 style={{ 
                    margin: 0, 
                    fontSize: '36px', 
                    fontWeight: '700',
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    WebkitBackgroundClip: 'text',
                    WebkitTextFillColor: 'transparent',
                    backgroundClip: 'text'
                  }}>
                    Albumi
                  </h1>
                  <p style={{ margin: '4px 0 0 0', color: '#666', fontSize: '14px' }}>
                    Otkrijte svoje omiljene albume
                  </p>
                </div>
              </div>
            </div>
            {isAdmin() && (
              <button 
                className="btn btn-primary" 
                onClick={() => setShowForm(!showForm)}
                style={{
                  padding: '14px 28px',
                  fontSize: '16px',
                  fontWeight: '600',
                  borderRadius: '12px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px'
                }}
              >
                {showForm ? '✕ Otkaži' : '➕ Dodaj album'}
              </button>
            )}
          </div>

          {/* Add/Edit Form */}
          {showForm && isAdmin() && (
            <div style={{
              marginTop: '30px',
              padding: '30px',
              background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%)',
              borderRadius: '16px',
              border: '2px solid rgba(102, 126, 234, 0.2)'
            }}>
              <h3 style={{ marginTop: 0, marginBottom: '24px', fontSize: '24px', fontWeight: '600', color: '#333' }}>
                {editingAlbum ? '✏️ Izmeni album' : '➕ Dodaj novi album'}
              </h3>
              <form onSubmit={handleSubmit}>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '16px' }}>
                  <div className="form-group">
                    <label>💿 Naziv albuma:</label>
                    <input
                      type="text"
                      name="name"
                      value={formData.name}
                      onChange={handleChange}
                      required
                      placeholder="Unesite naziv albuma"
                    />
                  </div>
                  <div className="form-group">
                    <label>📅 Datum izdavanja:</label>
                    <input
                      type="date"
                      name="releaseDate"
                      value={formData.releaseDate}
                      onChange={handleChange}
                    />
                  </div>
                </div>
                <div className="form-group">
                  <label>🎵 Žanr:</label>
                  <input
                    type="text"
                    name="genre"
                    value={formData.genre}
                    onChange={handleChange}
                    required
                    placeholder="Npr. Pop, Rock, Jazz"
                  />
                  <small style={{ display: 'block', marginTop: '6px', color: '#666', fontSize: '12px' }}>
                    Primer: Pop, Rock, Jazz, Electronic
                  </small>
                </div>
                <div className="form-group">
                  <label>🎤 Izvođači (držite Ctrl/Cmd za višestruki izbor):</label>
                  <select
                    name="selectedArtistIds"
                    multiple
                    value={formData.selectedArtistIds}
                    onChange={handleArtistSelect}
                    required
                    style={{ minHeight: '120px' }}
                  >
                    {artists.map((artist) => (
                      <option key={artist.id} value={artist.id}>
                        {artist.name}
                      </option>
                    ))}
                  </select>
                  {formData.selectedArtistIds.length > 0 && (
                    <small style={{ display: 'block', marginTop: '6px', color: '#667eea', fontWeight: '500' }}>
                      ✓ Izabrano: {formData.selectedArtistIds.length} izvođač(a)
                    </small>
                  )}
                </div>
                {error && (
                  <div className="error" style={{ marginBottom: '16px' }}>
                    <span>⚠️</span>
                    <span>{error}</span>
                  </div>
                )}
                <div style={{ display: 'flex', gap: '12px', marginTop: '20px' }}>
                  <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>
                    {editingAlbum ? '💾 Sačuvaj izmene' : '➕ Dodaj album'}
                  </button>
                  <button type="button" className="btn btn-secondary" onClick={handleCancel}>
                    ✕ Otkaži
                  </button>
                </div>
              </form>
            </div>
          )}

          {error && !showForm && (
            <div className="error" style={{ marginTop: '20px' }}>
              <span>⚠️</span>
              <span>{error}</span>
            </div>
          )}

          {/* Search and Filter Controls */}
          <div style={{ 
            marginTop: '30px', 
            padding: '24px', 
            background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.08) 0%, rgba(118, 75, 162, 0.08) 100%)',
            borderRadius: '16px',
            border: '1px solid rgba(102, 126, 234, 0.2)'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '20px' }}>
              <span style={{ fontSize: '24px' }}>🔍</span>
              <h3 style={{ margin: 0, fontSize: '20px', fontWeight: '600', color: '#333' }}>Pretraga i Filtriranje</h3>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '16px', marginBottom: '16px' }}>
              <div className="form-group" style={{ marginBottom: 0 }}>
                <label style={{ marginBottom: '8px' }}>🔎 Pretraga po nazivu:</label>
                <input
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Unesite naziv albuma..."
                />
              </div>
              <div className="form-group" style={{ marginBottom: 0 }}>
                <label style={{ marginBottom: '8px' }}>🎵 Filtriranje po žanru:</label>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <select
                    value={selectedGenre}
                    onChange={(e) => setSelectedGenre(e.target.value)}
                    style={{ flex: 1 }}
                  >
                    <option value="">🎵 Svi žanrovi</option>
                    {allGenres.length > 0 ? allGenres.map((genre) => (
                      <option key={genre} value={genre}>
                        {genre}
                      </option>
                    )) : (
                      <>
                        <option value="Pop">Pop</option>
                        <option value="Rock">Rock</option>
                        <option value="Jazz">Jazz</option>
                        <option value="Classical">Classical</option>
                        <option value="Electronic">Electronic</option>
                        <option value="Hip-Hop">Hip-Hop</option>
                        <option value="Country">Country</option>
                        <option value="R&B">R&B</option>
                        <option value="Reggae">Reggae</option>
                      </>
                    )}
                  </select>
                  {isAuthenticated && !isAdmin() && selectedGenre && (
                    <button
                      className={subscribedGenres.includes(selectedGenre) ? "btn btn-secondary" : "btn btn-primary"}
                      onClick={() => handleSubscribeToGenre(selectedGenre)}
                      disabled={isSubscribing}
                      style={{ 
                        whiteSpace: 'nowrap',
                        padding: '14px 20px',
                        fontSize: '16px'
                      }}
                      title={subscribedGenres.includes(selectedGenre) ? `Odjavite se sa žanra: ${selectedGenre}` : `Pretplati se na žanr: ${selectedGenre}`}
                    >
                      {isSubscribing 
                        ? '⏳' 
                        : (subscribedGenres.includes(selectedGenre) ? '✓ Pretplaćen' : '🔔 Pretplati se')}
                    </button>
                  )}
                </div>
              </div>
            </div>
            {subscriptionMessage && (
              <div className="success" style={{ marginTop: '12px', marginBottom: '12px' }}>
                <span>✓</span>
                <span>{subscriptionMessage}</span>
              </div>
            )}
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              gap: '8px',
              padding: '12px',
              background: 'rgba(102, 126, 234, 0.1)',
              borderRadius: '8px',
              fontSize: '14px',
              fontWeight: '500',
              color: '#667eea'
            }}>
              <span>📊</span>
              <span>Pronađeno albuma: <strong>{filteredAlbums.length}</strong> od <strong>{albums.length}</strong></span>
            </div>
          </div>
        </div>

        {/* Albums Grid */}
        <div style={{ marginTop: '30px' }}>
          {filteredAlbums.length === 0 ? (
            <div style={{
              background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
              backdropFilter: 'blur(10px)',
              borderRadius: '20px',
              padding: '60px 40px',
              textAlign: 'center',
              boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
              border: '1px solid rgba(255,255,255,0.3)'
            }}>
              <div style={{ fontSize: '64px', marginBottom: '20px' }}>💿</div>
              <h3 style={{ fontSize: '24px', fontWeight: '600', color: '#333', marginBottom: '12px' }}>
                {albums.length === 0 ? 'Nema albuma' : 'Nema albuma koji odgovaraju pretrazi'}
              </h3>
              <p style={{ color: '#666', fontSize: '16px' }}>
                {albums.length === 0 
                  ? 'Dodajte prvi album da biste počeli!' 
                  : 'Pokušajte da promenite filter ili pretragu da biste videli više rezultata.'}
              </p>
            </div>
          ) : (
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
              gap: '24px'
            }}>
              {filteredAlbums.map((album) => (
                <div
                  key={album.id}
                  style={{
                    background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
                    backdropFilter: 'blur(10px)',
                    borderRadius: '20px',
                    padding: '28px',
                    boxShadow: '0 10px 30px rgba(0,0,0,0.1)',
                    border: '1px solid rgba(255,255,255,0.3)',
                    cursor: 'pointer',
                    transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                    position: 'relative',
                    overflow: 'hidden'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.transform = 'translateY(-8px)';
                    e.currentTarget.style.boxShadow = '0 20px 50px rgba(102, 126, 234, 0.25)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'translateY(0)';
                    e.currentTarget.style.boxShadow = '0 10px 30px rgba(0,0,0,0.1)';
                  }}
                  onClick={() => navigate(`/albums/${album.id}`)}
                >
                  {/* Album Icon */}
                  <div style={{
                    width: '80px',
                    height: '80px',
                    borderRadius: '20px',
                    background: `linear-gradient(135deg, 
                      hsl(${(album.id.charCodeAt(0) * 137.508) % 360}, 70%, 60%) 0%, 
                      hsl(${(album.id.charCodeAt(0) * 137.508 + 60) % 360}, 70%, 50%) 100%)`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '40px',
                    marginBottom: '20px',
                    boxShadow: '0 8px 20px rgba(0,0,0,0.15)',
                    transition: 'transform 0.3s ease'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.transform = 'scale(1.1) rotate(5deg)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'scale(1) rotate(0deg)';
                  }}
                  >
                    💿
                  </div>

                  {/* Album Name */}
                  <h3 style={{
                    margin: '0 0 16px 0',
                    fontSize: '24px',
                    fontWeight: '700',
                    color: '#333',
                    lineHeight: '1.3'
                  }}>
                    {album.name}
                  </h3>

                  {/* Album Details */}
                  <div style={{ 
                    display: 'flex', 
                    flexDirection: 'column',
                    gap: '12px',
                    marginBottom: '16px',
                    padding: '12px',
                    background: 'rgba(102, 126, 234, 0.05)',
                    borderRadius: '8px'
                  }}>
                    {album.genre && (
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                        <span style={{ fontSize: '16px' }}>🎵</span>
                        <span 
                          className="genre-tag"
                          style={{
                            padding: '6px 14px',
                            background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
                            color: '#667eea',
                            border: '1px solid rgba(102, 126, 234, 0.2)',
                            fontWeight: '500',
                            fontSize: '12px',
                            borderRadius: '16px'
                          }}
                        >
                          {album.genre}
                        </span>
                      </div>
                    )}
                    {album.releaseDate && (
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <span style={{ fontSize: '16px' }}>📅</span>
                        <span style={{ fontSize: '14px', color: '#666', fontWeight: '500' }}>Datum izdavanja:</span>
                        <span style={{ fontSize: '14px', fontWeight: '600', color: '#333' }}>
                          {new Date(album.releaseDate).toLocaleDateString('sr-RS', {
                            year: 'numeric',
                            month: 'long',
                            day: 'numeric'
                          })}
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Admin Actions */}
                  {isAdmin() && (
                    <div style={{ 
                      display: 'flex', 
                      gap: '8px', 
                      marginTop: '20px',
                      paddingTop: '20px',
                      borderTop: '1px solid #e0e0e0'
                    }}>
                      <button
                        className="btn btn-secondary"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleEdit(album);
                        }}
                        style={{
                          flex: 1,
                          padding: '10px 16px',
                          fontSize: '14px',
                          fontWeight: '500'
                        }}
                      >
                        ✏️ Izmeni
                      </button>
                      <button
                        className="btn btn-danger"
                        onClick={async (e) => {
                          e.stopPropagation();
                          if (window.confirm(`Da li ste sigurni da želite da obrišete album "${album.name}"?`)) {
                            try {
                              await api.deleteAlbum(album.id);
                              loadAlbums();
                            } catch (err) {
                              setError(err.message || 'Greška pri brisanju albuma');
                            }
                          }
                        }}
                        style={{
                          flex: 1,
                          padding: '10px 16px',
                          fontSize: '14px',
                          fontWeight: '500'
                        }}
                      >
                        🗑️ Obriši
                      </button>
                    </div>
                  )}

                  {/* View Arrow */}
                  <div style={{
                    position: 'absolute',
                    top: '24px',
                    right: '24px',
                    fontSize: '20px',
                    opacity: 0.3,
                    transition: 'all 0.3s ease'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.opacity = 1;
                    e.currentTarget.style.transform = 'translateX(4px)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.opacity = 0.3;
                    e.currentTarget.style.transform = 'translateX(0)';
                  }}
                  >
                    →
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Albums;
