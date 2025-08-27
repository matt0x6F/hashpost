'use client';

import React, { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from './shadcn/dialog';
import { Button } from './shadcn/button';
import { Input } from './shadcn/input';
import { Label } from './shadcn/label';
import { Textarea } from './shadcn/textarea';
import { Checkbox } from './shadcn/checkbox';
import { Plus, Loader2, Search, X, User } from 'lucide-react';
import { useAuth } from '@/lib/auth-context';
import { getApi } from '@/lib/api-client';
import { SubforumsApi } from '@/generated/api/src/apis/SubforumsApi';
import { SearchApi } from '@/generated/api/src/apis/SearchApi';
import type { SubforumCreateBody } from '@/generated/api/src/models/SubforumCreateBody';
import type { PublicSearchPseudonym } from '@/generated/api/src/models/PublicSearchPseudonym';
import { toast } from 'sonner';
import { RulesEditor, type Rule } from './RulesEditor';

interface CreateForumDialogProps {
  onForumCreated?: (forumName: string) => void;
  children?: React.ReactNode;
}

export function CreateForumDialog({ onForumCreated, children }: CreateForumDialogProps) {
  const { user, isAuthenticated } = useAuth();
  const [open, setOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const [formData, setFormData] = useState<SubforumCreateBody>({
    name: '',
    slug: '',
    description: '',
    communityType: 't',
    isNsfw: false,
    isPrivate: false,
    isRestricted: false,
    rulesText: '',
    sidebarText: '',
    coModerators: [],
  });
  const [rules, setRules] = useState<Rule[]>([]);

  // Co-moderator search state
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<PublicSearchPseudonym[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [selectedCoModerators, setSelectedCoModerators] = useState<PublicSearchPseudonym[]>([]);

  const isDemocraticSubforum = formData.communityType === 't' || formData.communityType === 'g';

  const handleInputChange = (field: keyof SubforumCreateBody, value: string | boolean) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
    
    // Auto-generate slug from name
    if (field === 'name') {
      const slug = (value as string)
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');
      setFormData(prev => ({
        ...prev,
        name: value as string,
        slug,
      }));
    }

    // Reset co-moderators when community type changes
    if (field === 'communityType') {
      setSelectedCoModerators([]);
      setFormData(prev => ({
        ...prev,
        coModerators: [],
      }));
    }
  };

  const searchPseudonyms = async (query: string) => {
    if (!query.trim() || query.length < 2) {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    try {
      const searchApi = getApi(SearchApi);
      const response = await searchApi.searchPseudonymsPublic(query, 1, 25);
      setSearchResults(response.pseudonyms || []);
    } catch (error) {
      console.error('Error searching pseudonyms:', error);
      toast.error('Failed to search pseudonyms');
    } finally {
      setIsSearching(false);
    }
  };

  const addCoModerator = (pseudonym: PublicSearchPseudonym) => {
    // Check if already selected
    if (selectedCoModerators.some(cm => cm.pseudonymId === pseudonym.pseudonymId)) {
      toast.error('This pseudonym is already selected as a co-moderator');
      return;
    }

    // Check if it's the user's own pseudonym
    if (pseudonym.pseudonymId === user?.activePseudonymId) {
      toast.error('You cannot select your own pseudonym as a co-moderator');
      return;
    }

    const newCoModerators = [...selectedCoModerators, pseudonym];
    setSelectedCoModerators(newCoModerators);
    setFormData(prev => ({
      ...prev,
      coModerators: newCoModerators.map(cm => cm.pseudonymId),
    }));

    // Clear search
    setSearchQuery('');
    setSearchResults([]);
  };

  const removeCoModerator = (pseudonymId: string) => {
    const newCoModerators = selectedCoModerators.filter(cm => cm.pseudonymId !== pseudonymId);
    setSelectedCoModerators(newCoModerators);
    setFormData(prev => ({
      ...prev,
      coModerators: newCoModerators.map(cm => cm.pseudonymId),
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isAuthenticated) {
      setError('You must be logged in to create a forum');
      return;
    }

    // Validate required fields
    if (!formData.name.trim()) {
      setError('Forum name is required');
      return;
    }
    if (!formData.description.trim()) {
      setError('Description is required');
      return;
    }
    if (!formData.slug.trim()) {
      setError('Slug is required');
      return;
    }

    // Validate co-moderators for democratic subforums
    if (isDemocraticSubforum) {
      if (selectedCoModerators.length !== 2) {
        setError('Democratic subforums require exactly 2 co-moderators');
        return;
      }
    }

    setIsLoading(true);
    setError(null);

    try {
      const subforumsApi = getApi(SubforumsApi);
      
      // Convert structured rules to JSON string for backend compatibility
      let rulesText = '';
      if (rules.length > 0) {
        try {
          rulesText = JSON.stringify(rules);
        } catch {
          setIsLoading(false);
          setError('Failed to serialize rules. Please check your rules for invalid data.');
          return;
        }
      }
      
      const response = await subforumsApi.createSubforum({
        ...formData,
        rulesText: rulesText,
      });
      
      setOpen(false);
      setFormData({
        name: '',
        slug: '',
        description: '',
        communityType: 't',
        isNsfw: false,
        isPrivate: false,
        isRestricted: false,
        rulesText: '',
        sidebarText: '',
        coModerators: [],
      });
      setRules([]);
      setSelectedCoModerators([]);
      
      // Show success toast using Sonner
      toast.success(`Forum "${response.subforum.communityType}/${response.subforum.name}" created successfully!`, {
        description: `Your new forum is now live and ready for discussions.`,
        action: {
          label: "View Forum",
          onClick: () => {
            // TODO: Navigate to the forum page
          },
        },
      });
      
      if (onForumCreated) {
        onForumCreated(response.subforum.name);
      }
    } catch (err: unknown) {
      console.error('Error creating forum:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to create forum. Please try again.';
      setError(errorMessage);
      
      // Show error toast using Sonner
      toast.error('Failed to create forum', {
        description: errorMessage,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const canCreateForum = user?.capabilities?.includes('create_subforum');

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children || (
          <Button disabled={!isAuthenticated || !canCreateForum}>
            <Plus className="w-4 h-4 mr-2" />
            Create Forum
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Forum</DialogTitle>
        </DialogHeader>
        
        {!isAuthenticated ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground mb-4">You must be logged in to create a forum.</p>
            <Button onClick={() => setOpen(false)}>Close</Button>
          </div>
        ) : !canCreateForum ? (
          <div className="text-center py-8">
            <p className="text-muted-foreground mb-4">You don&apos;t have permission to create forums.</p>
            <Button onClick={() => setOpen(false)}>Close</Button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                <p className="text-sm text-destructive">{error}</p>
              </div>
            )}

            <div className="space-y-6">
              <div>
                <Label htmlFor="name" className="mb-1 block">Forum Name *</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('name', e.target.value)}
                  placeholder="Enter forum name"
                  maxLength={50}
                  required
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  This will be the display name of your forum
                </p>
              </div>

              <div>
                <Label htmlFor="slug" className="mb-1 block">URL Slug *</Label>
                <Input
                  id="slug"
                  value={formData.slug}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('slug', e.target.value)}
                  placeholder="forum-url-slug"
                  maxLength={30}
                  pattern="[a-z0-9-]+"
                  title="Only lowercase letters, numbers, and hyphens allowed"
                  required
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  This will be used in the URL: /{formData.communityType}/{formData.slug || 'forum-name'}
                </p>
              </div>

              <div>
                <Label htmlFor="description" className="mb-1 block">Description *</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => handleInputChange('description', e.target.value)}
                  placeholder="Describe what this forum is about"
                  maxLength={500}
                  rows={3}
                  required
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  {formData.description.length}/500 characters
                </p>
              </div>

              <div>
                <Label htmlFor="communityType" className="mb-1 block">Community Type *</Label>
                <select
                  id="communityType"
                  value={formData.communityType}
                  onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {
                    const newType = e.target.value;
                    setFormData(prev => ({
                      ...prev,
                      communityType: newType,
                    }));
                  }}
                  className="w-full h-10 px-3 py-2 border border-input bg-background rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                  required
                >
                  <option value="t">Topical (t/) - Generic subjects like programming, cooking (Democratic governance)</option>
                  <option value="g">Geographic (g/) - Location-based like Seattle, Tacoma (Democratic governance)</option>
                  <option value="b">Branded (b/) - Company/brand-owned like Apple, Minecraft (Owned governance)</option>
                  <option value="c">Creator (c/) - Individual creator-owned communities (Owned governance)</option>
                </select>
              </div>

              <div>
                <Label htmlFor="governanceStyle" className="mb-1 block">Governance Style</Label>
                <div className="w-full h-10 px-3 py-2 border border-input bg-background rounded-md text-sm flex items-center text-muted-foreground">
                  {isDemocraticSubforum ? (
                    <>
                      <span className="font-medium text-foreground">Democratic</span>
                      <span className="ml-2">- Owner commits to elected moderators</span>
                    </>
                  ) : (
                    <>
                      <span className="font-medium text-foreground">Owned</span>
                      <span className="ml-2">- Owner directly manages all moderators</span>
                    </>
                  )}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {isDemocraticSubforum 
                    ? 'Topical and Geographic communities use Democratic governance and require 2 co-moderators'
                    : 'Branded and Creator communities use Owned governance'
                  }
                </p>
              </div>

              {/* Co-Moderator Selection for Democratic Subforums */}
              {isDemocraticSubforum && (
                <div className="space-y-4">
                  <div>
                    <Label className="mb-2 block">Co-Moderators *</Label>
                    <p className="text-sm text-muted-foreground mb-3">
                      Democratic subforums require exactly 2 co-moderators. You cannot select your own pseudonym.
                    </p>
                    
                    {/* Selected Co-Moderators */}
                    {selectedCoModerators.length > 0 && (
                      <div className="space-y-2 mb-4">
                        <Label className="text-sm font-medium">Selected Co-Moderators ({selectedCoModerators.length}/2)</Label>
                        <div className="space-y-2">
                          {selectedCoModerators.map((coMod) => (
                            <div key={coMod.pseudonymId} className="flex items-center justify-between p-3 bg-muted rounded-md">
                              <div className="flex items-center gap-2">
                                <User className="w-4 h-4 text-muted-foreground" />
                                <span className="font-medium">{coMod.displayName}</span>
                                <span className="text-sm text-muted-foreground">@{coMod.slug}</span>
                              </div>
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => removeCoModerator(coMod.pseudonymId)}
                                className="h-8 w-8 p-0"
                              >
                                <X className="w-4 h-4" />
                              </Button>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Co-Moderator Search */}
                    {selectedCoModerators.length < 2 && (
                      <div className="space-y-3">
                        <div className="relative">
                          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                          <Input
                            type="text"
                            placeholder="Search for pseudonyms by display name or email..."
                            value={searchQuery}
                            onChange={(e) => {
                              const query = e.target.value;
                              setSearchQuery(query);
                              if (query.trim()) {
                                searchPseudonyms(query);
                              } else {
                                setSearchResults([]);
                              }
                            }}
                            className="pl-10"
                          />
                        </div>

                        {/* Search Results */}
                        {searchResults.length > 0 && (
                          <div className="border rounded-md max-h-48 overflow-y-auto">
                            {searchResults.map((result) => (
                              <button
                                key={result.pseudonymId}
                                type="button"
                                onClick={() => addCoModerator(result)}
                                className="w-full p-3 text-left hover:bg-muted border-b last:border-b-0 transition-colors"
                              >
                                <div className="flex items-center gap-2">
                                  <User className="w-4 h-4 text-muted-foreground" />
                                  <span className="font-medium">{result.displayName}</span>
                                  <span className="text-sm text-muted-foreground">@{result.slug}</span>
                                </div>
                              </button>
                            ))}
                          </div>
                        )}

                        {isSearching && (
                          <div className="flex items-center justify-center py-4">
                            <Loader2 className="w-4 h-4 animate-spin mr-2" />
                            <span className="text-sm text-muted-foreground">Searching...</span>
                          </div>
                        )}
                      </div>
                    )}

                    {/* Validation Message */}
                    {selectedCoModerators.length !== 2 && (
                      <p className="text-sm text-amber-600">
                        {selectedCoModerators.length === 0 
                          ? 'Please select 2 co-moderators to continue.'
                          : `Please select ${2 - selectedCoModerators.length} more co-moderator${2 - selectedCoModerators.length === 1 ? '' : 's'}.`
                        }
                      </p>
                    )}
                  </div>
                </div>
              )}

              <div className="space-y-3">
                <Label className="mb-1 block">Forum Settings</Label>
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="isNsfw"
                    checked={formData.isNsfw}
                    onCheckedChange={(checked: boolean | 'indeterminate') => handleInputChange('isNsfw', checked === true)}
                  />
                  <Label htmlFor="isNsfw" className="text-sm font-normal">
                    NSFW Content
                  </Label>
                </div>

                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="isPrivate"
                    checked={formData.isPrivate}
                    onCheckedChange={(checked: boolean | 'indeterminate') => handleInputChange('isPrivate', checked === true)}
                  />
                  <Label htmlFor="isPrivate" className="text-sm font-normal">
                    Private Forum (requires approval to join)
                  </Label>
                </div>

                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="isRestricted"
                    checked={formData.isRestricted}
                    onCheckedChange={(checked: boolean | 'indeterminate') => handleInputChange('isRestricted', checked === true)}
                  />
                  <Label htmlFor="isRestricted" className="text-sm font-normal">
                    Restricted (only approved users can post)
                  </Label>
                </div>
              </div>

              <div>
                <RulesEditor
                  rules={rules}
                  onChange={setRules}
                  disabled={isLoading}
                  showTitle={false}
                  maxRules={5}
                />
              </div>

              <div>
                <Label htmlFor="sidebarText" className="mb-1 block">Sidebar Content (Optional)</Label>
                <Textarea
                  id="sidebarText"
                  value={formData.sidebarText || ''}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => handleInputChange('sidebarText', e.target.value)}
                  placeholder="Additional information to display in the sidebar"
                  maxLength={1000}
                  rows={3}
                  className="mt-1"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  {(formData.sidebarText || '').length}/1000 characters
                </p>
              </div>
            </div>

            <div className="flex justify-end space-x-2 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
                disabled={isLoading}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isLoading || (isDemocraticSubforum && selectedCoModerators.length !== 2)}>
                {isLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Creating...
                  </>
                ) : (
                  'Create Forum'
                )}
              </Button>
            </div>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
} 