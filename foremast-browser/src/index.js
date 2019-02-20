import React from 'react';
import { Provider } from 'react-redux';
import { render } from 'react-dom';
import { Router, Route, Switch } from 'react-router-dom';

import App from './App';
import store from './store';
import history from './history';

render(
  <Provider store={store}>
      <Router history={history}>
        <Switch>
          <Route path="/:namespace/:appName" component={App} />
          <Route path="/" component={App} />
        </Switch>
    </Router>
  </Provider>,
  document.getElementById('root')
);

