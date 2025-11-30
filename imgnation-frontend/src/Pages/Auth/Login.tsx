import { useState } from 'react';
import { Mail, Lock, Eye, EyeOff, ArrowLeft } from 'lucide-react';
import { NavLink } from 'react-router';
import {useAuth} from "../../Providers/Api.tsx";
import {useNavigate} from "npm:react-router@7.6.2";

const LoginPage = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [isLoading, setIsLoading] = useState(false);

    const { login } = useAuth();
    const navigate = useNavigate()

    const handleSubmit = async () => {
        if (!email || !password) {
            alert('Please fill in both email and password');
            return;
        }
        setIsLoading(true);
        try {
            await login({email, password})
            await navigate('/me')
        } catch (error: unknown) {
            alert("Login failed, check the console log")
            console.error(error);
        } finally {
            setIsLoading(false);
        }
    };

    return (
      <div className="bg-gray-50 max-w-2xl m-auto flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 rounded-lg shadow-lg">
          <div className="max-w-md w-full space-y-8">
              <div>
                  <div className="flex items-center justify-between mb-8">
                      {/* Back Button */}
                      <NavLink
                        to="/"
                        className="group flex items-center space-x-2 text-gray-600 hover:text-indigo-600 transition-all duration-200"
                      >
                          <div className="flex items-center justify-center w-9 h-9 rounded-full bg-white shadow-sm group-hover:shadow-md group-hover:bg-indigo-50 transition-all duration-200">
                              <ArrowLeft className="h-4 w-4 group-hover:text-indigo-600" />
                          </div>
                          <span className="text-sm font-medium group-hover:text-indigo-600">
                                Home
                            </span>
                      </NavLink>

                      {/* Title */}
                      <h2 className="text-3xl font-extrabold text-gray-900">
                          Sign In
                      </h2>

                      {/* Spacer for balance */}
                      <div className="w-16"></div>
                  </div>
                  <p className="mt-2 text-center text-sm text-gray-600">
                      Or{' '}
                      <NavLink
                        className="font-medium text-indigo-600 hover:text-indigo-500 transition-colors"
                        to='/signup'
                      >
                          create a new account
                      </NavLink>
                  </p>
              </div>

              <div className="mt-8 space-y-6">
                  <div className="space-y-4">
                      {/* Email Field */}
                      <div>
                          <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                              Email address
                          </label>
                          <div className="relative">
                              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                  <Mail className="h-5 w-5 text-gray-400" />
                              </div>
                              <input
                                id="email"
                                name="email"
                                type="email"
                                autoComplete="email"
                                required
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                placeholder="Enter your email"
                              />
                          </div>
                      </div>

                      {/* Password Field */}
                      <div>
                          <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                              Password
                          </label>
                          <div className="relative">
                              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                  <Lock className="h-5 w-5 text-gray-400" />
                              </div>
                              <input
                                id="password"
                                name="password"
                                type={showPassword ? 'text' : 'password'}
                                autoComplete="current-password"
                                required
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="block w-full pl-10 pr-12 py-2 border border-gray-300 rounded-md placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                placeholder="Enter your password"
                              />
                              <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
                                  <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    className="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600"
                                  >
                                      {showPassword ? (
                                        <EyeOff className="h-5 w-5" />
                                      ) : (
                                        <Eye className="h-5 w-5" />
                                      )}
                                  </button>
                              </div>
                          </div>
                      </div>
                  </div>

                  <div className="mt-8 flex justify-center">
                      <button
                        type="button"
                        onClick={handleSubmit}
                        disabled={isLoading}
                        className="group relative w-48 flex justify-center py-4 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                      >
                          {isLoading ? (
                            <div className="flex items-center">
                                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                                Signing in...
                            </div>
                          ) : (
                            'Sign in'
                          )}
                      </button>
                  </div>
              </div>
          </div>
      </div>
    )
}
export default LoginPage