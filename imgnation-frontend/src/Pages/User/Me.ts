import { useAuth } from "../../Providers/Api.tsx";
import { useEffect } from "react";
import { useNavigate } from "react-router";

export const MyPage = () => {
  const { isAuth, internal } = useAuth()
  const { userId }  = internal;
  const navigate = useNavigate();

  useEffect(() => {
    navigate(isAuth ? `/users?id=${userId}`: `/signup`);
  }, []);
  return null;
}