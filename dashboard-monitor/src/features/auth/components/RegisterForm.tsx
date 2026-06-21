import { useState } from 'react'
import { Eye, EyeOff, Loader2, Phone, Lock, User } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useRegister } from '@/features/auth/hooks/useRegister'
import { SuccessModal } from './SuccessModal'

interface FieldErrors {
    userName?: string
    phoneNumber?: string
    password?: string
    confirmPassword?: string
}

function validateFields(
    userName: string,
    phoneNumber: string,
    password: string,
    confirmPassword: string
): FieldErrors {
    const errors: FieldErrors = {}

    if (!userName.trim()) {
        errors.userName = 'Full name is required.'
    }

    if (!phoneNumber.trim()) {
        errors.phoneNumber = 'Phone number is required.'
    } else if (!/^09\d{9}$/.test(phoneNumber.trim())) {
        errors.phoneNumber = 'Enter a valid phone number (e.g. 09xxxxxxxxx).'
    }

    if (!password) {
        errors.password = 'Password is required.'
    } else if (password.length < 6) {
        errors.password = 'Password must be at least 6 characters.'
    }

    if (!confirmPassword) {
        errors.confirmPassword = 'Please confirm your password.'
    } else if (password !== confirmPassword) {
        errors.confirmPassword = 'Passwords do not match.'
    }

    return errors
}

export const RegisterForm = () => {
    const { register, isLoading, error, successUserId } = useRegister()

    const [userName, setUserName] = useState('')
    const [phoneNumber, setPhoneNumber] = useState('')
    const [password, setPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')
    const [showPassword, setShowPassword] = useState(false)
    const [showConfirmPassword, setShowConfirmPassword] = useState(false)
    const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
    const [hasSubmitted, setHasSubmitted] = useState(false)

    const revalidate = (
        field: keyof FieldErrors,
        value: string
    ) => {
        if (!hasSubmitted) return
        const snapshot = { userName, phoneNumber, password, confirmPassword, [field]: value }
        setFieldErrors(
            validateFields(snapshot.userName, snapshot.phoneNumber, snapshot.password, snapshot.confirmPassword)
        )
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setHasSubmitted(true)
        const errors = validateFields(userName, phoneNumber, password, confirmPassword)
        setFieldErrors(errors)
        if (Object.keys(errors).length > 0) return
        await register({ user_name: userName.trim(), phone_number: phoneNumber.trim(), password })
    }

    const inputClass = (hasError: boolean) =>
        `w-full h-10 pl-9 pr-4 rounded-lg border bg-background text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed ${hasError ? 'border-destructive' : 'border-input'}`

    return (
        <>
            {successUserId !== null && <SuccessModal userId={successUserId} />}

            <div className="w-full">
                {/* Header */}
                <div className="mb-8 text-center">
                    <div className="inline-flex items-center justify-center w-12 h-12 rounded-2xl bg-primary/10 mb-4">
                        <User size={22} className="text-primary" />
                    </div>
                    <h1 className="text-2xl font-semibold text-foreground tracking-tight">
                        Create an account
                    </h1>
                    <p className="text-sm text-muted-foreground mt-1">
                        Join ActivityMonitor to access your dashboards
                    </p>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="space-y-4">
                    {/* Full Name */}
                    <div className="space-y-1.5">
                        <label htmlFor="userName" className="text-sm font-medium text-foreground">
                            Full Name
                        </label>
                        <div className="relative">
                            <User
                                size={15}
                                className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
                            />
                            <input
                                id="userName"
                                type="text"
                                value={userName}
                                onChange={(e) => {
                                    setUserName(e.target.value)
                                    revalidate('userName', e.target.value)
                                }}
                                placeholder="Your full name"
                                disabled={isLoading}
                                className={inputClass(!!fieldErrors.userName)}
                            />
                        </div>
                        {fieldErrors.userName && (
                            <p className="text-xs text-destructive">{fieldErrors.userName}</p>
                        )}
                    </div>

                    {/* Phone Number */}
                    <div className="space-y-1.5">
                        <label htmlFor="phoneNumber" className="text-sm font-medium text-foreground">
                            Phone Number
                        </label>
                        <div className="relative">
                            <Phone
                                size={15}
                                className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
                            />
                            <input
                                id="phoneNumber"
                                type="tel"
                                value={phoneNumber}
                                onChange={(e) => {
                                    setPhoneNumber(e.target.value)
                                    revalidate('phoneNumber', e.target.value)
                                }}
                                placeholder="09xxxxxxxxx"
                                disabled={isLoading}
                                className={inputClass(!!fieldErrors.phoneNumber)}
                            />
                        </div>
                        {fieldErrors.phoneNumber && (
                            <p className="text-xs text-destructive">{fieldErrors.phoneNumber}</p>
                        )}
                    </div>

                    {/* Password */}
                    <div className="space-y-1.5">
                        <label htmlFor="password" className="text-sm font-medium text-foreground">
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
                                onChange={(e) => {
                                    setPassword(e.target.value)
                                    revalidate('password', e.target.value)
                                }}
                                placeholder="Min. 6 characters"
                                disabled={isLoading}
                                className={`w-full h-10 pl-9 pr-10 rounded-lg border bg-background text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed ${fieldErrors.password ? 'border-destructive' : 'border-input'}`}
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
                        {fieldErrors.password && (
                            <p className="text-xs text-destructive">{fieldErrors.password}</p>
                        )}
                    </div>

                    {/* Confirm Password */}
                    <div className="space-y-1.5">
                        <label htmlFor="confirmPassword" className="text-sm font-medium text-foreground">
                            Confirm Password
                        </label>
                        <div className="relative">
                            <Lock
                                size={15}
                                className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
                            />
                            <input
                                id="confirmPassword"
                                type={showConfirmPassword ? 'text' : 'password'}
                                value={confirmPassword}
                                onChange={(e) => {
                                    setConfirmPassword(e.target.value)
                                    revalidate('confirmPassword', e.target.value)
                                }}
                                placeholder="Re-enter your password"
                                disabled={isLoading}
                                className={`w-full h-10 pl-9 pr-10 rounded-lg border bg-background text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all disabled:opacity-50 disabled:cursor-not-allowed ${fieldErrors.confirmPassword ? 'border-destructive' : 'border-input'}`}
                            />
                            <button
                                type="button"
                                onClick={() => setShowConfirmPassword((v) => !v)}
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                                tabIndex={-1}
                            >
                                {showConfirmPassword ? <EyeOff size={15} /> : <Eye size={15} />}
                            </button>
                        </div>
                        {fieldErrors.confirmPassword && (
                            <p className="text-xs text-destructive">{fieldErrors.confirmPassword}</p>
                        )}
                    </div>

                    {/* API Error */}
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
                                Creating account…
                            </>
                        ) : (
                            'Create account'
                        )}
                    </button>

                    <p className="text-center text-sm text-muted-foreground pt-2">
                        Already have an account?{' '}
                        <Link to="/login" className="text-primary hover:underline font-medium">
                            Sign in
                        </Link>
                    </p>
                </form>
            </div>
        </>
    )
}
