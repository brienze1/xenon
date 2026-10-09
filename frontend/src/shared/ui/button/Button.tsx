import type { ComponentProps } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './Button.module.css'

export type ButtonVariant = 'primary' | 'accent' | 'secondary' | 'ghost' | 'danger'
export type ButtonSize = 'sm' | 'md'

interface ButtonProps extends ComponentProps<'button'> {
  variant?: ButtonVariant
  size?: ButtonSize
}

export function Button({ variant = 'secondary', size = 'md', className, type = 'button', ...props }: ButtonProps) {
  return <button type={type} className={cn(styles.button, styles[variant], styles[size], className)} {...props} />
}
