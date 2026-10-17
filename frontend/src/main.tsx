import { ClerkProvider } from '@clerk/react';
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ClerkProvider publishableKey={import.meta.env.VITE_CLERK_PUBLISHABLE_KEY} afterSignOutUrl="/"
      appearance={{ variables: {
        colorPrimary: '#ff8069', colorPrimaryForeground: '#191511',
        colorBackground: '#19191c', colorForeground: '#f2f1ee',
        colorMutedForeground: '#969ba5', colorNeutral: '#ffffff',
        colorInput: '#141416', colorInputForeground: '#f2f1ee',
        colorBorder: '#363a40', borderRadius: '0.5rem',
        fontFamily: 'Inter, ui-sans-serif, system-ui, sans-serif',
      } }}
      localization={{ signIn: { start: { title: 'Welcome to Automata', subtitle: 'Sign in to pick up where you left off.' } }, signUp: { start: { title: 'Create your account', subtitle: 'Your next workflow starts here.' } } }}
    >
      <App />
    </ClerkProvider>
  </StrictMode>,
)
