import initialState from '../store/initialState';
import { UPDATE_HIGHLIGHT_TIMESTAMP, CLEAR_HIGHLIGHT_TIMESTAMP } from '../actions/actionTypes';

export default function highlight(state = initialState.highlight, action) {
  let newState;
  switch (action.type) {
    case UPDATE_HIGHLIGHT_TIMESTAMP:
      newState = {
        timestamp: action.timestamp,
      };
      return newState;
    case CLEAR_HIGHLIGHT_TIMESTAMP:
      newState = {
        timestamp: null,
      };
      return newState;
    default:
      return state;
  }
}