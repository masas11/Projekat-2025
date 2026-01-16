import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

const RecoverAccount = () => {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    try {
      await api.requestMagicLink(email);
      setSuccess('Ako email postoji, magic link je poslat na vašu email adresu. Kliknite na link da biste se prijavili.');
    } catch (err) {
      setError(err.message || 'Greška pri slanju magic link-a');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container">
      <div className="card">
        <h2>Povraćaj naloga - Magic Link</h2>
        <p style={{ marginBottom: '20px', color: '#666' }}>
          Unesite vašu email adresu i mi ćemo vam poslati magic link za prijavu bez lozinke.
        </p>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Email adresa:</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="unesite@email.com"
            />
          </div>
          {error && <div className="error">{error}</div>}
          {success && <div className="success">{success}</div>}
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Slanje...' : 'Pošalji magic link'}
          </button>
          <div style={{ marginTop: '15px', textAlign: 'center' }}>
            <Link to="/login">Nazad na prijavu</Link>
          </div>
        </form>
      </div>
    </div>
  );
};

export default RecoverAccount;
