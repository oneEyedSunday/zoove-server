import React from 'react';
import ReactDOM from 'react-dom';
import './assets/main.css';
import App from './App';
import * as serviceWorker from './serviceWorker';
import { BrowserRouter } from 'react-router-dom'
import { createBrowserHistory } from 'history'
import ReactGA from 'react-ga'

const trackingID = 'UA-178149278-1'
ReactGA.initialize(trackingID)

const history = createBrowserHistory()

history.listen(lx => {
  ReactGA.set({ page: lx.location.pathname })
  ReactGA.pageview(lx.location.pathname)
})
ReactDOM.render(
  <BrowserRouter history>
    <React.StrictMode>
      <App />
    </React.StrictMode>
  </BrowserRouter>,
  document.getElementById('root')
);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
