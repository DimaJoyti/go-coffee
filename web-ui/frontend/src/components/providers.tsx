// Providers component - simplified version for when dependencies are not installed
// This file provides basic provider functionality without external dependencies

interface ProvidersProps {
  children: any
}

export function Providers({ children }: ProvidersProps) {
  // Simple wrapper that just returns children
  // When dependencies are installed, this can be replaced with proper providers
  return children
}

// Export placeholder hooks for compatibility
export const useTheme = () => ({
  theme: 'dark',
  setTheme: () => {},
})

export const useQuery = () => ({
  isLoading: false,
  error: null,
  data: null,
  refetch: () => {},
})
