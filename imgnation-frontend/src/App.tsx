import './App.css'
import { Routes, Route } from "react-router"
import LoginPage from "./Pages/Auth/Login.tsx";
import SignupPage from "./Pages/Auth/SignUp.tsx";
import UploadCard from "./Pages/Upload/Upload.tsx";
import UploadViewPage from "./Pages/Uploads/Uploads.tsx"
import {MyPage} from "./Pages/User/Me.ts";
import UserPage from "./Pages/User/User.tsx";
import {FeedPage} from "./Pages/Feed.tsx";
import {useFeed} from "./hooks/useFeed.ts";

const App = () => {
  const feed = useFeed();
  return (
    <Routes>
      <Route index element={<FeedPage feed={feed} />} />
      <Route path="/feed" element={<FeedPage feed={feed} />} />
      <Route path="/upload" element={<UploadCard />} />
      <Route path="/login" element={<LoginPage/>} />
      <Route path="/signup" element={<SignupPage />} />
      <Route path="/uploads" element={<UploadViewPage/>} />
      <Route path="/me" element={<MyPage/>} />
      <Route path="/users" element={<UserPage/>} />
    </Routes>
  )
}

export default App