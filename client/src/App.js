import React from 'react';
import logo from './logo.svg';
import { Switch, Route } from 'react-router-dom'
import './App.css';
import Login from './views/Login';
import Signup from './views/Signup';
import DeezerAuth from './views/auth/Deezer';
import Landing from './views/Landing';

function App() {
  return (
    <Switch>
      <Route path="/" component={Landing} />
      <Route path="/login" component={Login} />
      <Route path="/signup" component={Signup} />
      <Route path="/deezer/oauth" exact component={DeezerAuth} />
    </Switch>
  );
}

export default App;
