// just a page that says "Please check your email for verification link"
// this page shpuld alaso have a resend lkink button
// if clicked, send request to server to resend link
// server checks if verification token already generated for user within 24 hours, if yes, 
// send same link again


import React, { useState } from "react";
import { useLocation } from "react-router-dom";
import axios from "axios";
import "./ResendEmail.css";

export function ResendEmail() {
  const location = useLocation();
  const email = location.state?.email || "";
  const username = location.state?.username || "";
  const link = location.state?.link || "";

  const [message, setMessage] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
 

  

  const handleResend = async () => {
    setIsLoading(true);
    setMessage("");
    

  
    
    try {
      const res = await axios.post("http://localhost:8080/resend-verification", {
        email: email,
        username: username
      });
      
      setIsSuccess(true);
      setMessage("Verification link has been resent to your email!");
      
      console.log("New link:", res.data.link); // delete this 
      
    } catch (error) {
      setIsSuccess(false);
      
      if (error.response) {
        setMessage(error.response.data || "Failed to resend verification link.");
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
    <div className="verification-container">
      <div className="verification-card">
        <div className="email-icon">📧</div>
        
        <h2>Verify Your Email</h2>
        
        <p className="verification-message">
          We've sent a verification link to
        </p>
        
        <p className="email-address">{email}</p>
        
        <p className="instruction">
          Please check your email and click the verification link to activate your account.
        </p>
        
        <div className="divider"></div>
        
        <p className="resend-text">
          Didn't receive the email?
        </p>
        
        <button 
          onClick={handleResend} 
          disabled={isLoading}
          className="resend-button"
        >
          {isLoading ? "Sending..." : "Resend Verification Link"}
        </button>
        
        {message && (
          <p className={`message ${isSuccess ? "success" : "error"}`}>
            {message}
          </p>
        )}
        
        {/* Development only - shows the link */}
        {process.env.NODE_ENV === 'development' && link && (
          <div className="dev-link">
            <p><strong>Development Mode:</strong></p>
            <a href={link} target="_blank" rel="noopener noreferrer">
              Click here to verify
            </a>
          </div>
        )}
      </div>
    </div>
  );
}

