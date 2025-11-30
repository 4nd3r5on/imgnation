import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router';
import {useAuth} from "../../Providers/Api.tsx";

interface UserPublicData {
  id: string;
  profile_pic_file_id: string;
  username: string;
  name: string;
  roles: string[];
  updated_at: string;
  created_at: string;
}

const UserPage = () => {
  const urlParams = new URLSearchParams(globalThis.location.search);
  const id = urlParams.get("id")
  const username = urlParams.get("username")

  const [userData, setUserData] = useState<UserPublicData | null>(null)

  const navigate = useNavigate()
  const { req, logout } = useAuth()


  useEffect(() => {
    if (id != null) {
      req(`/users/${id}`, {
        method: 'GET',
      }).then((resp: Response) => resp.json())
        .then((data: UserPublicData) => {
          setUserData(data)
        })
    } else if (userData != null) {
      req(`/users/username/${username}`, {
        method: 'GET',
      }).then((resp: Response) => resp.json())
        .then((data: UserPublicData) => {
          setUserData(data)
        })
    }

  }, [])

  return userData ? (
    <div className="flex items-center justify-between bg-gradient-to-r from-blue-50 to-indigo-100 p-4 rounded-lg shadow-md border border-blue-200">
      <div className="flex items-center space-x-3">
        <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-full flex items-center justify-center text-white font-semibold text-lg">
          {userData.username.charAt(0).toUpperCase()}
        </div>
        <div>
          <p className="text-sm text-gray-600 font-medium">Welcome back,</p>
          <p className="text-lg font-bold text-gray-800">{userData.username}</p>
        </div>
      </div>
      <button
        type="button"
        onClick={() => { logout(); navigate('/')}}
        className="px-6 py-2 text-red-600 bg-red-200 font-semibold rounded-lg shadow-sm border border-blue-200 hover:bg-blue-50 hover:shadow-md transition-all duration-200 ease-in-out transform hover:scale-105"
      >
        Logout
      </button>
      <button
        type="button"
        onClick={() => navigate('/')}
        className="px-6 py-2 bg-white text-blue-600 font-semibold rounded-lg shadow-sm border border-blue-200 hover:bg-blue-50 hover:shadow-md transition-all duration-200 ease-in-out transform hover:scale-105"
      >
        🏠 Home
      </button>
    </div>
  ) : (
    <div className="flex items-center justify-center p-8">
      <div className="flex items-center space-x-3">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <p className="text-gray-600 font-medium">Loading your profile...</p>
      </div>
    </div>
  )
}

export default UserPage