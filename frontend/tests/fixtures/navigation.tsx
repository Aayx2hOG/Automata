// Isolated navigation harness; all API responses are supplied by Playwright.
import ReactDOM from 'react-dom/client';
import { Workspace } from '../../src/App';
import { AuthContext } from '../../src/context/auth-context';
import '../../src/index.css';
ReactDOM.createRoot(document.getElementById('root')!).render(
  <AuthContext.Provider value={{ user: null, isAuthenticated: false, isLoading: false, error: null, clearError: () => {}, logout: async () => {} }}><Workspace /></AuthContext.Provider>,
);
