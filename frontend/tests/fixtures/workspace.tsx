// Browser-only harness. API requests still require real, server-issued JWTs.
// Loaded by Playwright interception; never imported by the application entry point.
import ReactDOM from 'react-dom/client';
import { Workspace } from '../../src/App';
import { AuthContext } from '../../src/context/auth-context';
import { api } from '../../src/lib/api';
import '../../src/index.css';

const user = await api.auth.me();
ReactDOM.createRoot(document.getElementById('root')!).render(
  <AuthContext.Provider value={{ user, isAuthenticated: true, isLoading: false, error: null, clearError: () => {}, logout: async () => {} }}>
    <Workspace />
  </AuthContext.Provider>,
);
