# Frontend Conventions (React/TypeScript)

## Tech Stack

- Framework: React 19 with TanStack Router
- Styling: Tailwind CSS v4
- Icons: HugeIcons
- Build: Vite
- Testing: Vitest

## Project Structure

```
fe/src/
├── components/    # Reusable components
├── routes/        # Route components
├── lib/           # Utilities, API clients
├── types/         # TypeScript types
└── styles/        # Global styles
```

## Code Style

| Rule | Convention |
|------|------------|
| Semicolons | Never |
| Quotes | Single quotes |
| Trailing commas | Always |
| Formatting | Prettier handles it |

## Naming Conventions

| Type | Pattern | Example |
|------|---------|---------|
| Components | PascalCase | `UserCard`, `ExpenseList` |
| Functions | camelCase | `getUserById`, `calculateSplit` |
| Variables | camelCase | `currentUser`, `expenseAmount` |
| Types/Interfaces | PascalCase | `User`, `Expense` |
| Constants | SCREAMING_SNAKE_CASE | `API_BASE_URL` |

## Component Pattern

```tsx
interface UserCardProps {
  user: User
  onEdit?: (user: User) => void
}

export function UserCard({ user, onEdit }: UserCardProps) {
  return (
    <div className="p-4 border rounded-lg">
      <h3 className="text-lg font-semibold">{user.name}</h3>
      <p className="text-gray-600">{user.email}</p>
    </div>
  )
}
```

## Imports Order

```tsx
// React imports
import { useState } from 'react'

// Third-party imports
import { useQuery } from '@tanstack/react-query'

// Internal imports
import { api } from '~/lib/api'
import { UserCard } from '~/components/UserCard'
```

## Anti-Patterns

- Using `any` type
- Direct API calls in components (use hooks/services)
- Inline styles (use Tailwind)
- Magic strings/numbers (use constants)
