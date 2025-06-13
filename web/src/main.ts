import './style.css';
import { Dashboard } from './dashboard.js';

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  console.log('🚀 SDL Canvas Dashboard starting...');
  new Dashboard();
});