import axios from 'axios';
import { getAccessToken } from './AuthService';

const BASE_URL = 'http://localhost:3333';

export {getAppsUsage, getTypingSpeed};

function getAppsUsage() {
    const url = `${BASE_URL}/apps/usage`;
    return axios.get(url).then(response => response.data);
}

function getTypingSpeed() {
    const url = `${BASE_URL}/typing/speed`;
    return axios.get(url).then(response => response.data);
}