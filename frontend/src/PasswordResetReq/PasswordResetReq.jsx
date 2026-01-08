// just a page that say enter your email
// sends emailid to backend
// after etting reponse, displays email sent is user exists
// redirects to enter the token sent to email page

import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import "./PasswordResetReq.css";

export function PasswordResetReq() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!email) {
      setMessage("Please enter your email");
      setIsSuccess(false);
      return;
    }

    setIsLoading(true);
    setMessage("");

    try {
      const res = await axios.post("http://localhost:8080/reset-password-request", {
        email: email
      });

      setIsSuccess(true);
      setMessage("If an account exists with this email, a reset code has been sent");

      // Redirect to password reset page with email
      setTimeout(() => {
        navigate("/reset-password", { state: { email } });
      }, 2000);

    } catch (error) {
      setIsSuccess(false);

      if (error.response) {
        setMessage(error.response.data || "Failed to send reset code");
      } else if (error.request) {
        setMessage("Cannot connect to server. Please try again later.");
      } else {
        setMessage("An error occurred. Please try again.");
      }

      console.error("Error:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="password-reset-container">
      <div className="password-reset-card">
        <div className="lock-icon">🔒</div>
        
        <h2>Forgot Password?</h2>
        
        <p className="reset-instruction">
          Enter your email address 
        </p>

        <form onSubmit={handleSubmit}>
          <div className="input-group">
            <input
              type="email"
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isLoading}
              required
            />
          </div>

          <button type="submit" disabled={isLoading} className="submit-button">
            {isLoading ? "Sending..." : "Send Reset Code"}
          </button>

          {message && (
            <p className={`message ${isSuccess ? "success" : "error"}`}>
              {message}
            </p>
          )}
        </form>

        <div className="back-to-login">
          <p>
            Remember your password?{" "}
            <a href="/login" onClick={(e) => {
              e.preventDefault();
              navigate("/login");
            }}>
              Back to Login
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}