import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';

const Register = () => {
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    username: '',
    password: '',
    confirmPassword: '',
  });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [verificationLink, setVerificationLink] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

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

    if (formData.password !== formData.confirmPassword) {
      setError('Lozinke se ne poklapaju');
      return;
    }

    const passwordError = validatePassword(formData.password);
    if (passwordError) {
      setError(passwordError);
      return;
    }

    setLoading(true);
    setVerificationLink('');

    try {
      const response = await api.register(formData);
      
      // Check if verification link is in response (when SMTP is not configured)
      if (response.verificationLink) {
        setVerificationLink(response.verificationLink);
        setSuccess('Uspešna registracija! SMTP nije konfigurisan, koristite link ispod za verifikaciju:');
      } else {
        setSuccess('Uspešna registracija! Email za verifikaciju je poslat. Proverite svoj email i kliknite na link za verifikaciju.');
      }
    } catch (err) {
      // Better error handling for 409 Conflict (user already exists)
      if (err.message && (err.message.includes('409') || err.message.includes('already exists') || err.message.includes('user already exists'))) {
        setError('Korisnik sa ovim korisničkim imenom ili email adresom već postoji. Molimo koristite drugačije podatke ili se prijavite.');
      } else {
        setError(err.message || 'Greška pri registraciji');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container">
      <div className="card">
        <h2>Registruj se</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Ime:</label>
            <input
              type="text"
              name="firstName"
              value={formData.firstName}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label>Prezime:</label>
            <input
              type="text"
              name="lastName"
              value={formData.lastName}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label>Email:</label>
            <input
              type="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label>Korisničko ime:</label>
            <input
              type="text"
              name="username"
              value={formData.username}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label>Lozinka:</label>
            <input
              type="password"
              name="password"
              value={formData.password}
              onChange={handleChange}
              required
            />
            <small style={{ color: '#666', fontSize: '12px' }}>
              Lozinka mora imati najmanje 8 karaktera, jedno veliko slovo i jedan broj
            </small>
          </div>
          <div className="form-group">
            <label>Potvrdi lozinku:</label>
            <input
              type="password"
              name="confirmPassword"
              value={formData.confirmPassword}
              onChange={handleChange}
              required
            />
          </div>
          {error && <div className="error">{error}</div>}
          {success && (
            <div className="success">
              <p>{success}</p>
              {verificationLink && (
                <div style={{ marginTop: '15px', padding: '15px', backgroundColor: '#f0f0f0', borderRadius: '5px' }}>
                  <p style={{ marginBottom: '10px', fontWeight: 'bold' }}>Verifikacioni link:</p>
                  <a 
                    href={verificationLink} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    style={{ 
                      color: '#FF9800', 
                      wordBreak: 'break-all',
                      textDecoration: 'underline',
                      display: 'block',
                      marginBottom: '10px'
                    }}
                  >
                    {verificationLink}
                  </a>
                  <button
                    onClick={() => {
                      navigator.clipboard.writeText(verificationLink);
                      alert('Link je kopiran u clipboard!');
                    }}
                    style={{
                      padding: '8px 15px',
                      backgroundColor: '#FF9800',
                      color: 'white',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: 'pointer',
                      fontSize: '14px'
                    }}
                  >
                    Kopiraj link
                  </button>
                </div>
              )}
            </div>
          )}
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Registracija...' : 'Registruj se'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default Register;
