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
            {/* Header */}
            <div className="mb-8 text-center">
                <div className="inline-flex items-center justify-center w-12 h-12 rounded-2xl bg-primary/10 mb-4">
                    <span className="text-2xl">📊</span>
                </div>
                <h1 className="text-2xl font-semibold text-foreground tracking-tight">
                    Welcome back
                </h1>
                <p className="text-sm text-muted-foreground mt-1">
                    Sign in to your ActivityMonitor account
                </p>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-4">
                {/* Phone Number */}
                <div className="space-y-1.5">
                    <label
                        htmlFor="phone"
                        className="text-sm font-medium text-foreground"
                    >
                        Phone Number
                    </label>
                    <div className="relative">
                        <Phone
                            size={15}
                            className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
                        />
                        <input
                            id="phone"
                            type="tel"
                            value={phoneNumber}
                            onChange={(e) => setPhoneNumber(e.target.value)}
                            placeholder="09xxxxxxxxx"
                            disabled={isLoading}
                            className="w-full h-10 pl-9 pr-4 rounded-lg border border-input bg-background text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                {/* Password */}
                <div className="space-y-1.5">
                    <label
                        htmlFor="password"
                        className="text-sm font-medium text-foreground"
                    >
                        Password
                    </label>
                    <div className="relative">
                        <Lock
                            size={15}
                            className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
                        />
                        <input
                            id="password"
                            type={showPassword ? 'text' : 'password'}
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="Enter your password"
                            disabled={isLoading}
                            className="w-full h-10 pl-9 pr-10 rounded-lg border border-input bg-background text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        />
                        <button
                            type="button"
                            onClick={() => setShowPassword((v) => !v)}
                            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                            tabIndex={-1}
                        >
                            {showPassword ? <EyeOff size={15} /> : <Eye size={15} />}
                        </button>
                    </div>
                </div>

                {/* Error message */}
                {error && (
                    <div className="flex items-center gap-2 px-3 py-2.5 rounded-lg bg-destructive/10 border border-destructive/20">
                        <span className="text-xs text-destructive font-medium">{error}</span>
                    </div>
                )}

                {/* Submit */}
                <button
                    type="submit"
                    disabled={isLoading}
                    className="w-full h-10 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 mt-2"
                >
                    {isLoading ? (
                        <>
                            <Loader2 size={15} className="animate-spin" />
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
