import {combineReducers} from 'redux';
import metric from './metricReducer';

const rootReducer = combineReducers({
  metric
});

export default rootReducer;