import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import {SignUp } from './SignUp/SignUp';
import {ResendEmail} from './ResendEmail/ResendEmail';
import { VerifyEmail } from './VerifyEmail/VerifyEmail';
import { Login } from './Login/LogIn'; 
import { PasswordResetReq } from './PasswordResetReq/PasswordResetReq';
import { ResetPassword } from './ResetPassword/ResetPassword';

import './App.css';


function App() {

  return (
    <Router>
      <div className="App">
        <Routes>
          <Route path="/" element={<SignUp />} />
          {/* <Route path="/signup" element={<SignUp />} /> */}
          <Route path="/verify" element={<VerifyEmail/>}/>
          <Route path="/resend" element={<ResendEmail/>} />
          <Route path="/login" element={<Login />} />
          <Route path="/pass-reset-req" element={<PasswordResetReq />} />
          <Route path="/reset-password" element={<ResetPassword/>} />

        </Routes>
      </div>
    </Router>
  );
}

export default App;