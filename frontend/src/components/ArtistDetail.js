import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api from '../services/api';

const ArtistDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [artist, setArtist] = useState(null);
  const [albums, setAlbums] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadArtist();
    loadAlbums();
  }, [id]);

  const loadArtist = async () => {
    try {
      const data = await api.getArtist(id);
      setArtist(data);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju izvođača');
    } finally {
      setLoading(false);
    }
  };

  const loadAlbums = async () => {
    try {
      const data = await api.getAlbumsByArtist(id);
      setAlbums(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error loading albums:', err);
    }
  };

  if (loading) {
    return <div className="container">Učitavanje...</div>;
  }

  if (error || !artist) {
    return (
      <div className="container">
        <div className="card">
          <div className="error">{error || 'Izvođač nije pronađen'}</div>
          <button className="btn btn-secondary" onClick={() => navigate('/artists')}>
            Nazad
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="card">
        <button className="btn btn-secondary" onClick={() => navigate('/artists')} style={{ marginBottom: '20px' }}>
          ← Nazad na izvođače
        </button>
        <h2>{artist.name}</h2>
        {artist.biography && <p style={{ marginTop: '10px', marginBottom: '10px' }}>{artist.biography}</p>}
        {artist.genres && artist.genres.length > 0 && (
          <div style={{ marginTop: '10px', marginBottom: '20px' }}>
            <strong>Žanrovi: </strong>
            {artist.genres.map((genre, idx) => (
              <span key={idx} className="genre-tag">{genre}</span>
            ))}
          </div>
        )}
      </div>

      {albums.length > 0 && (
        <div className="card">
          <h3>Albumi</h3>
          {albums.map((album) => (
            <div
              key={album.id}
              className="list-item"
              onClick={() => navigate(`/albums/${album.id}`)}
            >
              <h4>{album.name}</h4>
              {album.genre && <span className="genre-tag">{album.genre}</span>}
              {album.releaseDate && <p style={{ marginTop: '5px' }}>Datum izdavanja: {new Date(album.releaseDate).toLocaleDateString()}</p>}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ArtistDetail;
