import { ReactNode } from "react"
import { AuthProvider } from "./Api.tsx";
import {FeedProvider} from "./Feed.tsx";

const Providers = ({ children}: { children: ReactNode }) => {
  return (
    <AuthProvider>
      <FeedProvider>
        { children }
      </FeedProvider>
    </AuthProvider>
  )
}

export default Providers