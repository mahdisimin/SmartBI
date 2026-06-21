import { useState } from 'react'
import { Eye, EyeOff, Loader2, Phone, Lock } from 'lucide-react'
import { useLogin } from '@/features/auth/hooks/useLogin'

export const LoginForm = () => {
    const { login, isLoading, error } = useLogin()
    const [phoneNumber, setPhoneNumber] = useState('')
    const [password, setPassword] = useState('')
    const [showPassword, setShowPassword] = useState(false)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        await login({ phone_number: phoneNumber, password })
    }

    return (
        <div className="w-full">
            <div className="mb-8">
                <h1 className="text-2xl font-bold text-foreground tracking-tight mb-2">
                    Sign in
                </h1>
                <p className="text-sm text-muted-foreground">
                    Enter your credentials to access the platform
                </p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
                {/* Phone Number */}
                <div className="space-y-1.5">
                    <label htmlFor="phone" className="text-sm font-medium text-foreground">
                        Phone Number
                    </label>
                    <div className="relative">
                        <Phone
                            size={14}
                            className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                        />
                        <input
                            id="phone"
                            type="tel"
                            value={phoneNumber}
                            onChange={(e) => setPhoneNumber(e.target.value)}
                            placeholder="09xxxxxxxxx"
                            disabled={isLoading}
                            className="w-full h-10 pl-10 pr-4 rounded-lg border border-input bg-muted text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary/50 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                {/* Password */}
                <div className="space-y-1.5">
                    <label htmlFor="password" className="text-sm font-medium text-foreground">
                        Password
                    </label>
                    <div className="relative">
                        <Lock
                            size={14}
                            className="absolute left-3.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
                        />
                        <input
                            id="password"
                            type={showPassword ? 'text' : 'password'}
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="Enter your password"
                            disabled={isLoading}
                            className="w-full h-10 pl-10 pr-10 rounded-lg border border-input bg-muted text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary/50 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        />
                        <button
                            type="button"
                            onClick={() => setShowPassword((v) => !v)}
                            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                            tabIndex={-1}
                        >
                            {showPassword ? <EyeOff size={14} /> : <Eye size={14} />}
                        </button>
                    </div>
                </div>

                {/* Error */}
                {error && (
                    <div className="flex items-start gap-2.5 px-3.5 py-3 rounded-lg bg-destructive/10 border border-destructive/20">
                        <span className="text-xs text-destructive leading-relaxed">{error}</span>
                    </div>
                )}

                {/* Submit */}
                <button
                    type="submit"
                    disabled={isLoading || !phoneNumber || !password}
                    className="w-full h-10 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:ring-offset-2 focus:ring-offset-background transition-all disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center gap-2 mt-2"
                >
                    {isLoading ? (
                        <>
                            <Loader2 size={14} className="animate-spin" />
                            Signing in…
                        </>
                    ) : (
                        'Sign in'
                    )}
                </button>
            </form>
        </div>
    )
}
