import * as types from './actionTypes';

export const updateHighlightTimestamp = timestamp => ({
  type: types.UPDATE_HIGHLIGHT_TIMESTAMP,
  timestamp
});

export const clearHighlightTimestamp = () => ({
  type: types.CLEAR_HIGHLIGHT_TIMESTAMP,
});