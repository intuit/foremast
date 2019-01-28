import { combineReducers } from 'redux';
import highlight from './highlightReducer';

const rootReducer = combineReducers({
  highlight
});

export default rootReducer;