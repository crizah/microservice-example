import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import axios from "axios";
import "./VerifyEmail.css";


// user clicks on link
// frontend sends token to server

export function VerifyEmail() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState("verifying");
  const [message, setMessage] = useState("Verifying your email...");

  useEffect(() => {
    const verifyEmail = async () => {
      const token = searchParams.get("token");
      
      if (!token) {
        setStatus("error");
        setMessage("Invalid verification link.");
        return;
      }

      try {
        const res = await axios.get(`http://localhost:8080/verify-email?token=${token}`);
        
        setStatus("success");
        setMessage("Email verified successfully! Redirecting to login...");
        
        setTimeout(() => {
          navigate("/login");
        }, 2000);
        
      } catch (error) {
        setStatus("error");
        
        if (error.response?.status === 401) {
          setMessage("Invalid or expired verification link.");
        } else {
          setMessage("Verification failed. Please try again or request a new link.");
        }
      }
    };

    verifyEmail();
  }, [searchParams, navigate]);

  return (
    <div className="verify-email-container">
      <div className="verify-card">
        {status === "verifying" && (
          <>
            <div className="spinner">⏳</div>
            <h2>{message}</h2>
          </>
        )}
        
        {status === "success" && (
          <>
            <div className="success-icon">✓</div>
            <h2>{message}</h2>
          </>
        )}
        
        {status === "error" && (
          <>
            <div className="error-icon">✗</div>
            <h2>{message}</h2>
            <button onClick={() => navigate("/signup")}>
              Back to Sign Up
            </button>
          </>
        )}
      </div>
    </div>
  );
}
