import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import Login from './Login';

// Mock fetch globally
beforeEach(() => {
  global.fetch = jest.fn(() => Promise.resolve({
    ok: true,
    json: () => Promise.resolve({ url: 'https://example.com' })
  })) as any;
});

afterEach(() => {
  (global.fetch as jest.Mock).mockRestore();
});

test('renders provider buttons and calls API on click', async () => {
  render(<Login />);
  const googleBtn = screen.getByRole('button', { name: /sign in with google/i });
  expect(googleBtn).toBeInTheDocument();
  fireEvent.click(googleBtn);
  expect(global.fetch).toHaveBeenCalledWith('/api/v1/auth/login?provider=google');
});