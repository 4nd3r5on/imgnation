import { createContext, useContext, useState, useEffect, ReactNode, useCallback, useRef } from 'react';

const API_URL = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';

const AUTH_REFRESH_TOKEN_KEY = 'auth_refresh_token';
const AUTH_USERID_KEY = 'auth_user_id';

// Types
export interface SignUpReqData {
  username: string;
  name: string;
  email: string;
  password: string;
}

export interface LoginReqData {
  email: string;
  password: string;
}

export interface AuthRespData {
  access_token: string;
  refresh_token: string;
  user_id: string;
}

export interface RefreshAccessTokenRespData {
  access_token: string;
}

interface States {
  apiUrl: string;
  userId: string | null;
  accessToken: string | null;
  refreshToken: string | null;
}

interface AuthContextType {
  // State
  internal: States;
  isAuth: boolean;
  isInitialized: boolean;
  isLoading: boolean;

  // Methods
  req: (route: string, req: RequestInit) => Promise<Response>;
  login: (req: LoginReqData) => Promise<void>;
  signUp: (req: SignUpReqData) => Promise<void>;
  refreshAccessToken: () => Promise<void>;
  logout: () => void;
}

// Utility functions
const clearAuthStorage = (): void => {
  localStorage.removeItem(AUTH_REFRESH_TOKEN_KEY);
  localStorage.removeItem(AUTH_USERID_KEY);
};

const saveAuthToStorage = (data: AuthRespData): void => {
  localStorage.setItem(AUTH_REFRESH_TOKEN_KEY, data.refresh_token);
  localStorage.setItem(AUTH_USERID_KEY, data.user_id);
};

const getStoredAuth = (): { refreshToken: string | null; userId: string | null } => ({
  refreshToken: localStorage.getItem(AUTH_REFRESH_TOKEN_KEY),
  userId: localStorage.getItem(AUTH_USERID_KEY),
});

// API functions
const baseReq = async (
  endpoint: string,
  init?: RequestInit,
  noFailCodes?: Set<number>
): Promise<Response> => {
  try {
    const resp = await fetch(endpoint, init);

    if (
      (resp.status >= 200 && resp.status < 300) ||
      (noFailCodes?.has(resp.status))
    ) {
      return resp;
    }

    const contentType = resp.headers.get('Content-Type');
    const errorMessage = contentType?.includes('application/json')
      ? JSON.stringify(await resp.json())
      : await resp.text();

    throw new Error(
      `Request to ${endpoint} finished with non-ok status code (code ${resp.status}): ${errorMessage}`
    );
  } catch (error) {
    if (error instanceof Error) {
      throw error;
    }
    throw new Error(`failed to fetch ${endpoint}`);
  }
};

const loginReq = async (
  url: string,
  req: RequestInit | null,
  reqData: LoginReqData
): Promise<AuthRespData> => {
  return await (await baseReq(`${url}/auth/login`, {
    ...req,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...req?.headers,
    },
    body: JSON.stringify(reqData),
  })).json();
};

const signUpReq = async (
  url: string,
  req: RequestInit | null,
  reqData: SignUpReqData
): Promise<AuthRespData> => {
  return await (await baseReq(`${url}/auth/signup`, {
    ...req,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...req?.headers,
    },
    body: JSON.stringify(reqData),
  })).json();
};

const refreshAccessTokenReq = async (
  url: string,
  req: RequestInit | null,
  refreshToken: string | null
): Promise<string | null> => {
  if (!refreshToken) return null;

  try {
    const { access_token } = await (await baseReq(`${url}/auth/refresh`, {
      ...req,
      method: 'GET',
      headers: {
        Authorization: `Bearer ${refreshToken}`,
        ...req?.headers,
      },
    })).json() as RefreshAccessTokenRespData;

    return access_token;
  } catch {
    return null;
  }
};

// Context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [apiUrl] = useState(API_URL);
  const [authState, setAuthState] = useState<States>({
    apiUrl: API_URL,
    userId: null,
    accessToken: null,
    refreshToken: null,
  });

  const [isInitialized, setIsInitialized] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  // Use refs to prevent race conditions
  const initPromiseRef = useRef<Promise<void> | null>(null);
  const refreshPromiseRef = useRef<Promise<void> | null>(null);

  // Initialize authentication state
  useEffect(() => {
    const initAuth = async (): Promise<void> => {
      try {
        setIsLoading(true);
        const { refreshToken, userId } = getStoredAuth();

        if (refreshToken && userId) {
          const newAccessToken = await refreshAccessTokenReq(apiUrl, {}, refreshToken);

          if (newAccessToken) {
            setAuthState({
              apiUrl,
              userId,
              accessToken: newAccessToken,
              refreshToken,
            });
          } else {
            // Token refresh failed, clear storage
            clearAuthStorage();
          }
        }
      } catch (error) {
        console.error('Auth initialization failed:', error);
        clearAuthStorage();
      } finally {
        setIsInitialized(true);
        setIsLoading(false);
      }
    };

    initPromiseRef.current = initAuth();
  }, [apiUrl]);

  const saveAuthData = useCallback((data: AuthRespData): void => {
    saveAuthToStorage(data);
    setAuthState({
      apiUrl,
      userId: data.user_id,
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
    });
  }, [apiUrl]);

  const logout = useCallback((): void => {
    clearAuthStorage();
    setAuthState({
      apiUrl,
      userId: null,
      accessToken: null,
      refreshToken: null,
    });
    refreshPromiseRef.current = null;
  }, [apiUrl]);

  const refreshAccessToken = useCallback(async (): Promise<void> => {
    // Return existing promise if refresh is in progress
    if (refreshPromiseRef.current) {
      return refreshPromiseRef.current;
    }

    if (!authState.refreshToken) {
      logout();
      return;
    }

    const refreshPromise = (async (): Promise<void> => {
      try {
        const newAccessToken = await refreshAccessTokenReq(apiUrl, {}, authState.refreshToken);

        if (newAccessToken) {
          setAuthState(prev => ({
            ...prev,
            accessToken: newAccessToken,
          }));
        } else {
          logout();
        }
      } catch (error) {
        console.error('Failed to refresh access token:', error);
        logout();
      } finally {
        refreshPromiseRef.current = null;
      }
    })();

    refreshPromiseRef.current = refreshPromise;
    return refreshPromise;
  }, [authState.refreshToken, apiUrl, logout]);

  const req = useCallback(async (
    route: string,
    reqInit: RequestInit,
    tryRefreshingToken = true
  ): Promise<Response> => {
    // Wait for initialization
    if (initPromiseRef.current) {
      await initPromiseRef.current;
    }
    // Wait for any ongoing refresh
    if (refreshPromiseRef.current) {
      await refreshPromiseRef.current;
    }

    if (!route.startsWith('/')) {
      route = `/${route}`;
    }

    const resp = await baseReq(`${apiUrl}${route}`, {
      ...reqInit,
      headers: {
        ...reqInit.headers,
        ...(authState.accessToken && {
          Authorization: `Bearer ${authState.accessToken}`
        }),
      },
    }, new Set([401]));

    // Handle 401 with token refresh
    if (resp.status === 401 && tryRefreshingToken && authState.refreshToken) {
      await refreshAccessToken();
      return req(route, reqInit, false); // Retry without refresh
    }

    return resp;
  }, [apiUrl, authState.accessToken, authState.refreshToken, refreshAccessToken]);

  const login = useCallback(async (reqData: LoginReqData): Promise<void> => {
    const data = await loginReq(apiUrl, {}, reqData);
    saveAuthData(data);
  }, [apiUrl, saveAuthData]);

  const signUp = useCallback(async (reqData: SignUpReqData): Promise<void> => {
    const data = await signUpReq(apiUrl, {}, reqData);
    saveAuthData(data);
  }, [apiUrl, saveAuthData]);

  const contextValue: AuthContextType = {
    isInitialized,
    isLoading,
    internal: authState,
    isAuth: !!authState.accessToken && isInitialized,
    login,
    signUp,
    logout,
    refreshAccessToken,
    req,
  };

  // Show loading spinner while initializing
  if (!isInitialized) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-sm w-full mx-4">
          <div className="flex flex-col items-center space-y-4">
            <div className="w-12 h-12 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin" />
            <div className="text-center">
              <h2 className="text-lg font-semibold text-gray-800 mb-2">Loading</h2>
              <p className="text-gray-600 text-sm">Initializing authentication...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};