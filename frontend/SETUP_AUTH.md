# Authentication Setup Guide

This guide walks you through setting up Supabase authentication with GitHub OAuth for the Workflow application.

## Prerequisites

- Node.js 18+ installed
- A GitHub account
- Access to create a new Supabase project

## 1. Supabase Project Setup

### Create a New Supabase Project

1. Go to [Supabase Dashboard](https://supabase.com/dashboard)
2. Click "New Project"
3. Choose your organization
4. Enter project details:
   - **Name**: `workflow-app` (or your preferred name)
   - **Database Password**: Generate a strong password and save it
   - **Region**: Choose the closest region to your users
5. Click "Create new project"
6. Wait for the project to be created (usually takes 1-2 minutes)

### Enable GitHub Authentication Provider

1. In your Supabase dashboard, go to **Authentication** > **Providers**
2. Find **GitHub** in the list and click on it
3. Enable GitHub by toggling the switch
4. Leave **Client ID** and **Client Secret** empty for now (we'll fill these in step 3)
5. Click **Save**

### Configure Auth Settings

1. Go to **Authentication** > **Settings**
2. In **General settings**:
   - Set **Site URL**: `http://localhost:3000` (for development)
   - Add **Redirect URLs**:
     - `http://localhost:3000/auth/callback`
     - `https://your-domain.com/auth/callback` (for production)
3. Click **Save**

### Get Supabase Credentials

1. Go to **Settings** > **API**
2. Copy the following values (you'll need these for step 4):
   - **Project URL** (NEXT_PUBLIC_SUPABASE_URL)
   - **Project API keys** > **anon public** (NEXT_PUBLIC_SUPABASE_ANON_KEY)
   - **Project API keys** > **service_role** (SUPABASE_SERVICE_ROLE_KEY)

## 2. Database Schema Setup

### Create User Profiles Table

1. Go to **SQL Editor** in your Supabase dashboard
2. Click **New Query** and run the following SQL:

```sql
-- Create profiles table
CREATE TABLE public.profiles (
    id UUID REFERENCES auth.users(id) ON DELETE CASCADE PRIMARY KEY,
    email TEXT NOT NULL,
    full_name TEXT,
    avatar_url TEXT,
    username TEXT UNIQUE,
    github_username TEXT,
    bio TEXT,
    website TEXT,
    location TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Create workflows table
CREATE TABLE public.workflows (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'completed', 'archived')),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Create tasks table
CREATE TABLE public.tasks (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workflow_id UUID REFERENCES public.workflows(id) ON DELETE CASCADE NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'todo' CHECK (status IN ('todo', 'in_progress', 'completed')),
    priority TEXT DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    due_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Enable RLS (Row Level Security)
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.workflows ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.tasks ENABLE ROW LEVEL SECURITY;

-- Create policies
CREATE POLICY "Users can view own profile" ON public.profiles
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON public.profiles
    FOR UPDATE USING (auth.uid() = id);

CREATE POLICY "Users can view own workflows" ON public.workflows
    FOR ALL USING (auth.uid() = user_id);

CREATE POLICY "Users can view tasks from own workflows" ON public.tasks
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM public.workflows 
            WHERE workflows.id = tasks.workflow_id 
            AND workflows.user_id = auth.uid()
        )
    );

-- Create function to automatically create profile on signup
CREATE OR REPLACE FUNCTION public.handle_new_user() 
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.profiles (id, email, full_name, avatar_url, github_username)
    VALUES (
        NEW.id,
        NEW.email,
        NEW.raw_user_meta_data->>'full_name',
        NEW.raw_user_meta_data->>'avatar_url',
        NEW.raw_user_meta_data->>'user_name'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create trigger to call the function on new user signup
CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE PROCEDURE public.handle_new_user();
```

3. Click **Run** to execute the SQL

## 3. GitHub OAuth App Setup

### Create GitHub OAuth App

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click **OAuth Apps** in the left sidebar
3. Click **New OAuth App**
4. Fill in the application details:
   - **Application name**: `Workflow App` (or your preferred name)
   - **Homepage URL**: `http://localhost:3000` (for development)
   - **Application description**: `A workflow management application`
   - **Authorization callback URL**: Get this from your Supabase dashboard:
     - Go to **Authentication** > **Providers** > **GitHub**
     - Copy the **Callback URL** (it should look like: `https://your-project-id.supabase.co/auth/v1/callback`)
5. Click **Register application**

### Get GitHub Credentials

1. After creating the app, you'll see the **Client ID** - copy this
2. Click **Generate a new client secret** and copy the generated secret
3. Keep these values secure - you'll need them for step 4

### Update Supabase with GitHub Credentials

1. Go back to your Supabase dashboard
2. Navigate to **Authentication** > **Providers** > **GitHub**
3. Enter the **Client ID** and **Client Secret** from GitHub
4. Click **Save**

## 4. Environment Variables Setup

### Create Environment File

1. In your project root (`/frontend/`), copy the example environment file:
   ```bash
   cp .env.example .env.local
   ```

2. Edit `.env.local` and fill in your credentials:

```env
# Supabase Configuration
NEXT_PUBLIC_SUPABASE_URL=https://your-project-id.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=your_supabase_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key

# GitHub OAuth Configuration
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret

# Application Configuration
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=your_nextauth_secret_here
```

### Generate NEXTAUTH_SECRET

Run this command to generate a random secret:
```bash
openssl rand -base64 32
```

## 5. Application Integration

### Add Auth Provider to App

Update your `src/app/layout.tsx` to include the AuthProvider:

```tsx
import { AuthProvider } from '@/contexts/AuthContext'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          {children}
        </AuthProvider>
      </body>
    </html>
  )
}
```

### Install Additional Dependencies (if needed)

The required `@supabase/supabase-js` package is already installed. If you need testing dependencies:

```bash
npm install --save-dev @testing-library/react @testing-library/jest-dom jest jest-environment-jsdom
```

## 6. Testing the Setup

### Start Development Server

```bash
npm run dev
```

### Test Authentication Flow

1. Navigate to `http://localhost:3000/login`
2. Click "Continue with GitHub"
3. You should be redirected to GitHub for authorization
4. After authorization, you should be redirected back to your app
5. Check the browser's developer console for any errors

### Verify Database

1. In Supabase dashboard, go to **Table Editor**
2. Check that a new profile was created in the `profiles` table after signing in

## 7. Production Deployment

### Update Environment Variables

For production deployment, update the following:

1. **Supabase Settings**:
   - Add your production domain to **Redirect URLs**
   - Update **Site URL** to your production domain

2. **GitHub OAuth App**:
   - Update **Homepage URL** to your production domain
   - Update **Authorization callback URL** if using a different Supabase project

3. **Environment Variables**:
   - Update `NEXTAUTH_URL` to your production domain
   - Ensure all secrets are secure and different from development

### Security Checklist

- [ ] All environment variables are set correctly
- [ ] Row Level Security (RLS) is enabled on all tables
- [ ] Service role key is not exposed to the client
- [ ] GitHub OAuth app is configured with correct callback URLs
- [ ] NEXTAUTH_SECRET is a strong, random string

## 8. Troubleshooting

### Common Issues

1. **"Invalid redirect URL"**
   - Ensure the callback URL in GitHub matches exactly with Supabase
   - Check that redirect URLs are added to Supabase settings

2. **"Network request failed"**
   - Verify Supabase URL and keys are correct
   - Check that your Supabase project is active

3. **TypeScript errors**
   - Run `npm run build` to check for any compilation errors
   - Ensure all imports are correctly resolved

4. **Authentication not persisting**
   - Check browser console for session storage issues
   - Verify that `persistSession` is set to `true` in client config

### Getting Help

- [Supabase Documentation](https://supabase.com/docs)
- [GitHub OAuth Documentation](https://docs.github.com/en/developers/apps/building-oauth-apps)
- [Next.js Documentation](https://nextjs.org/docs)

## File Structure Created

```
src/
├── lib/
│   └── supabase/
│       ├── client.ts          # Supabase client initialization
│       └── auth.ts            # Authentication helper functions
├── types/
│   ├── auth.ts                # Authentication type definitions
│   └── database.ts            # Database schema types
├── contexts/
│   └── AuthContext.tsx        # React context for auth state
├── app/
│   ├── login/
│   │   └── page.tsx          # Login page component
│   └── auth/
│       └── callback/
│           └── page.tsx       # OAuth callback handler
└── __tests__/
    ├── lib/supabase/
    │   └── auth.test.ts       # Auth function tests
    └── contexts/
        └── AuthContext.test.tsx # Auth context tests
```

Your authentication system is now ready to use!