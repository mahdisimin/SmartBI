import { MessageSquare } from 'lucide-react'

export default function AiChatPage() {
    return (
        <div className="flex flex-col items-center justify-center py-24 text-center">
            <div className="w-12 h-12 rounded-2xl bg-muted flex items-center justify-center mb-4">
                <MessageSquare size={20} className="text-muted-foreground" />
            </div>
            <p className="text-sm font-medium text-foreground">AI Chat — Coming soon</p>
            <p className="text-xs text-muted-foreground mt-1">
                This feature is under development.
            </p>
        </div>
    )
}
