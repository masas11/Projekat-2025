import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const ChangePassword = () => {
  const { user } = useAuth();
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const validatePassword = (password) => {
    if (password.length < 8) {
      return 'Lozinka mora imati najmanje 8 karaktera';
    }
    if (!/[A-Z]/.test(password)) {
      return 'Lozinka mora sadržati najmanje jedno veliko slovo';
    }
    if (!/[0-9]/.test(password)) {
      return 'Lozinka mora sadržati najmanje jedan broj';
    }
    return null;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (newPassword !== confirmPassword) {
      setError('Nove lozinke se ne poklapaju');
      return;
    }

    const passwordError = validatePassword(newPassword);
    if (passwordError) {
      setError(passwordError);
      return;
    }

    setLoading(true);

    try {
      await api.changePassword({
        username: user.username,
        oldPassword,
        newPassword,
      });
      setSuccess('Lozinka je uspešno promenjena!');
      setTimeout(() => {
        navigate('/');
      }, 2000);
    } catch (err) {
      setError(err.message || 'Greška pri promeni lozinke');
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return (
      <div className="container">
        <div className="card">
          <div className="error">Morate biti prijavljeni da biste promenili lozinku.</div>
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="card">
        <h2>Promena lozinke</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Stara lozinka:</label>
            <input
              type="password"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <label>Nova lozinka:</label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
            />
            <small style={{ color: '#666', fontSize: '12px' }}>
              Lozinka mora imati najmanje 8 karaktera, jedno veliko slovo i jedan broj.
              Možete promeniti lozinku tek nakon što je prošao 1 dan od poslednje promene.
            </small>
          </div>
          <div className="form-group">
            <label>Potvrdi novu lozinku:</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
          </div>
          {error && <div className="error">{error}</div>}
          {success && <div className="success">{success}</div>}
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Promena...' : 'Promeni lozinku'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default ChangePassword;
