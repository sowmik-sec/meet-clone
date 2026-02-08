'use client';

import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useToast } from '@/components/ui/use-toast';
import { userApi } from '@/lib/api/user';

interface ProfileFormValues {
    name: string;
    bio: string;
}

export default function ProfileSettingsPage() {
    const { user } = useAuth();
    const { toast } = useToast();
    const [loading, setLoading] = useState(false);

    const { register, handleSubmit, setValue, formState: { errors } } = useForm<ProfileFormValues>({
        defaultValues: {
            name: '',
            bio: '',
        },
    });

    useEffect(() => {
        if (user) {
            setValue('name', user.name);
            setValue('bio', user.bio || '');
        }
    }, [user, setValue]);

    const onSubmit = async (data: ProfileFormValues) => {
        setLoading(true);
        try {
            const updatedUser = await userApi.updateProfile(data);

            // We need to update the local user state if useAuth doesn't automatically refetch
            // For now, let's assume a page reload or simple success message is enough
            // Ideally, update the context

            toast({
                title: "Profile updated",
                description: "Your profile details have been saved.",
            });

            // Reload window to refresh auth state if needed, or if we had a method to update auth context
            // window.location.reload(); 
        } catch (error) {
            console.error('Failed to update profile:', error);
            toast({
                title: "Error",
                description: "Failed to update profile. Please try again.",
                variant: "destructive",
            });
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-2xl mx-auto px-4 py-8">
            <Card>
                <CardHeader>
                    <CardTitle>Profile Settings</CardTitle>
                    <CardDescription>
                        Manage your public profile information.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                        <div className="space-y-2">
                            <Label htmlFor="name">Display Name</Label>
                            <Input
                                id="name"
                                placeholder="Your Name"
                                {...register('name', { required: 'Name is required', minLength: { value: 2, message: 'Name must be at least 2 characters' } })}
                            />
                            {errors.name && (
                                <p className="text-sm text-red-500">{errors.name.message}</p>
                            )}
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="bio">Bio</Label>
                            <Textarea
                                id="bio"
                                placeholder="Tell people a little about yourself..."
                                className="min-h-[100px]"
                                {...register('bio')}
                            />
                            <p className="text-xs text-gray-500">
                                This will be displayed on your public booking page.
                            </p>
                        </div>

                        <div className="flex justify-end">
                            <Button type="submit" disabled={loading}>
                                {loading ? 'Saving...' : 'Save Changes'}
                            </Button>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}
