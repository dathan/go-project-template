import React from 'react';
import { Routes, Route, Link } from 'react-router-dom';
import Login from './pages/Login';

const Home: React.FC = () => {
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Welcome</h1>
      <p>This is the home page. Please <Link to="/login" className="text-blue-600 underline">log in</Link> to continue.</p>
    </div>
  );
};

const App: React.FC = () => {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/login" element={<Login />} />
    </Routes>
  );
};

export default App;