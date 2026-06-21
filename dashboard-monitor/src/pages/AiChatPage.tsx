import { Sparkles } from 'lucide-react'

export default function AiChatPage() {
    return (
        <div className="p-8 min-h-full flex flex-col items-center justify-center py-24">
            <div className="text-center max-w-sm">
                <div className="w-14 h-14 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center mb-5 mx-auto">
                    <Sparkles size={22} className="text-primary" />
                </div>
                <p className="text-[11px] font-semibold text-primary uppercase tracking-widest mb-3">
                    Coming Soon
                </p>
                <p className="text-base font-bold text-foreground tracking-tight mb-2">
                    AI Analytics Assistant
                </p>
                <p className="text-sm text-muted-foreground leading-relaxed">
                    Ask questions about your data, explore user behavior patterns, and get actionable insights — powered by AI.
                </p>
            </div>
        </div>
    )
}
