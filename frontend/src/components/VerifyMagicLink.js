import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const VerifyMagicLink = () => {
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState('verifying'); // verifying, success, error
  const [message, setMessage] = useState('');
  const navigate = useNavigate();
  const { login } = useAuth();

  useEffect(() => {
    const token = searchParams.get('token');
    const rawUrl = window.location.href;
    
    console.log('VerifyMagicLink - Raw URL:', rawUrl);
    console.log('VerifyMagicLink - Token from URL (decoded):', token);
    
    if (!token) {
      setStatus('error');
      setMessage('Token za magic link nije pronađen u URL-u. Proverite link.');
      return;
    }

    const verifyMagicLink = async () => {
      try {
        console.log('VerifyMagicLink - Calling API with token:', token);
        const response = await api.verifyMagicLink(token);
        // Magic link login automatski prijavljuje korisnika
        if (response.token && response.id) {
          login(response, response.token);
          setStatus('success');
          setMessage('Uspešno ste se prijavili pomoću magic link-a! Preusmeravanje...');
          setTimeout(() => {
            navigate('/');
          }, 2000);
        } else {
          setStatus('error');
          setMessage('Greška pri prijavljivanju pomoću magic link-a.');
        }
      } catch (err) {
        console.error('VerifyMagicLink - Error:', err);
        setStatus('error');
        setMessage(err.message || 'Greška pri verifikaciji magic link-a. Proverite da je link validan i nije istekao.');
      }
    };

    verifyMagicLink();
  }, [searchParams, navigate, login]);

  return (
    <div className="container">
      <div className="card">
        <h2>Povraćaj naloga - Magic Link</h2>
        {status === 'verifying' && (
          <div>
            <p>Verifikacija magic link-a u toku...</p>
          </div>
        )}
        {status === 'success' && (
          <div className="success">
            <p>{message}</p>
            <p>Preusmeravanje...</p>
          </div>
        )}
        {status === 'error' && (
          <div className="error">
            <p>{message}</p>
            <button
              className="btn btn-primary"
              onClick={() => navigate('/login')}
              style={{ marginTop: '10px' }}
            >
              Idi na prijavu
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default VerifyMagicLink;
