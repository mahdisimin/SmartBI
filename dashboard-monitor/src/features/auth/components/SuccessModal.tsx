import { CheckCircle } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

interface SuccessModalProps {
    userId: number
}

export const SuccessModal = ({ userId }: SuccessModalProps) => {
    const navigate = useNavigate()

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="animate-modal-in w-full max-w-sm mx-4 bg-card border border-border rounded-2xl p-8 shadow-2xl">
                <div className="flex justify-center mb-6">
                    <div className="w-16 h-16 rounded-full bg-emerald-500/10 flex items-center justify-center">
                        <CheckCircle size={32} className="text-emerald-500" />
                    </div>
                </div>

                <div className="text-center mb-6">
                    <h2 className="text-lg font-semibold text-foreground mb-2">
                        Registration Successful
                    </h2>
                    <p className="text-sm text-muted-foreground">
                        Your account has been created successfully.
                    </p>
                </div>

                <div className="bg-muted rounded-xl p-4 mb-6 text-center border border-border">
                    <p className="text-xs font-medium text-muted-foreground mb-1.5 uppercase tracking-widest">
                        User Number
                    </p>
                    <p className="text-3xl font-bold text-foreground font-mono tracking-tight">
                        {userId}
                    </p>
                </div>

                <button
                    onClick={() => navigate('/login', { replace: true })}
                    className="w-full h-10 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 transition-all"
                >
                    Go to Login
                </button>
            </div>
        </div>
    )
}
