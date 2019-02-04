import React from 'react';

import './Header.css';
import logo from './foremast-logo-white.png';

export default class Header extends React.Component {
  render() {
    return (
      <div className='header'>
        <img src={logo} alt=''/>
        <h1>Foremast Dashboard</h1>
      </div>
    );
  }
}