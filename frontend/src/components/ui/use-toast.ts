// Simplified toast hook
import { useState } from "react"

export const useToast = () => {
    const [toasts, setToasts] = useState<any[]>([])

    const toast = ({ title, description, variant }: any) => {
        console.log(`Toast: ${title} - ${description} (${variant})`)
        // In a real app, we'd add to state and render a Toaster component
        alert(`${title}\n${description}`)
    }

    return { toast }
}
