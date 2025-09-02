"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/shadcn/card";
import { Button } from "@/components/shadcn/button";
import { Input } from "@/components/shadcn/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/shadcn/table";
import { Badge } from "@/components/shadcn/badge";
import { Label } from "@/components/shadcn/label";
import { Users, Search, Eye, Edit, MoreHorizontal } from "lucide-react";
import { getApi } from "@/lib/api-client";
import { AdminApi } from "@/generated/api/src/apis/AdminApi";
import { SearchApi } from "@/generated/api/src/apis/SearchApi";
import { AdminPseudonymInfo } from "@/generated/api/src/models/AdminPseudonymInfo";
import { SearchPseudonym } from "@/generated/api/src/models/SearchPseudonym";
import { toast } from "sonner";

export function PseudonymListTab() {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  // Pseudonym search state
  const [pseudonyms, setPseudonyms] = useState<SearchPseudonym[]>([]);
  const [pseudonymSearchQuery, setPseudonymSearchQuery] = useState("");
  const [isPseudonymSearchLoading, setIsPseudonymSearchLoading] = useState(false);
  
  // Pseudonyms state
  const [allPseudonyms, setAllPseudonyms] = useState<AdminPseudonymInfo[]>([]);
  const [isPseudonymsLoading, setIsPseudonymsLoading] = useState(false);
  const [pseudonymsPage, setPseudonymsPage] = useState(1);
  const [pseudonymsTotalPages, setPseudonymsTotalPages] = useState(1);
  const [totalPseudonyms, setTotalPseudonyms] = useState(0);
  
  // Restore state from URL params
  useEffect(() => {
    const queryFromUrl = searchParams.get('query');
    if (queryFromUrl && queryFromUrl !== pseudonymSearchQuery) {
      setPseudonymSearchQuery(queryFromUrl);
      // Auto-search if query is in URL
      if (queryFromUrl.trim()) {
        searchPseudonymsWithQuery(queryFromUrl);
      }
    }

    // Handle return query from detail pages
    const returnQueryFromUrl = searchParams.get('returnQuery');
    if (returnQueryFromUrl && returnQueryFromUrl !== pseudonymSearchQuery) {
      setPseudonymSearchQuery(returnQueryFromUrl);
      // Auto-search if query is in URL
      if (returnQueryFromUrl.trim()) {
        searchPseudonymsWithQuery(returnQueryFromUrl);
      }
    }

    const pseudonymsPageFromUrl = searchParams.get('pseudonymsPage');
    if (pseudonymsPageFromUrl) {
      const page = parseInt(pseudonymsPageFromUrl);
      if (!isNaN(page) && page > 0) {
        setPseudonymsPage(page);
      }
    }

    // Handle return page from detail pages
    const returnPageFromUrl = searchParams.get('returnPage');
    if (returnPageFromUrl) {
      const page = parseInt(returnPageFromUrl);
      if (!isNaN(page) && page > 0) {
        setPseudonymsPage(page);
      }
    }
  }, [searchParams, pseudonymSearchQuery]);

  // Load pseudonyms when component mounts
  useEffect(() => {
    loadAllPseudonyms();
  }, [pseudonymsPage]);

  const loadAllPseudonyms = async () => {
    setIsPseudonymsLoading(true);
    try {
      const api = getApi(AdminApi);
      const response = await api.adminListPseudonyms(undefined, undefined, pseudonymsPage, 25);
      
      if (response.pseudonyms) {
        setAllPseudonyms(response.pseudonyms);
        setTotalPseudonyms(response.pagination?.total || 0);
        setPseudonymsTotalPages(response.pagination?.pages || 1);
      } else {
        setAllPseudonyms([]);
        setTotalPseudonyms(0);
        setPseudonymsTotalPages(1);
      }
    } catch (error: unknown) {
      console.error('Failed to load pseudonyms:', error);
      
      // Handle specific error cases
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'status' in error.response) {
        const status = (error.response as { status: number }).status;
        if (status === 401) {
          toast.error("Authentication required. Please log in again.");
        } else if (status === 403) {
          toast.error("Insufficient permissions for pseudonym management");
        } else if (status === 500) {
          toast.error("Server error. Please try again later.");
        } else {
          toast.error("Failed to load pseudonyms. Please try again.");
        }
      } else {
        toast.error("Failed to load pseudonyms. Please try again.");
      }
      
      setAllPseudonyms([]);
      setTotalPseudonyms(0);
      setPseudonymsTotalPages(1);
    } finally {
      setIsPseudonymsLoading(false);
    }
  };

  const searchPseudonymsWithQuery = async (query: string) => {
    if (!query.trim()) return;
    
    setIsPseudonymSearchLoading(true);
    try {
      const api = getApi(SearchApi);
      const response = await api.searchPseudonyms(query, undefined, undefined, 1, 100);
      
      if (response.pseudonyms) {
        setPseudonyms(response.pseudonyms);
        setTotalPseudonyms(response.pagination?.total || 0);
        setPseudonymsTotalPages(response.pagination?.pages || 1);
      } else {
        setPseudonyms([]);
        setTotalPseudonyms(0);
        setPseudonymsTotalPages(1);
      }
    } catch (error: unknown) {
      console.error('Failed to search pseudonyms:', error);
      toast.error("Failed to search pseudonyms. Please try again.");
      setPseudonyms([]);
      setTotalPseudonyms(0);
      setPseudonymsTotalPages(1);
    } finally {
      setIsPseudonymSearchLoading(false);
    }
  };

  const handlePseudonymSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!pseudonymSearchQuery.trim()) return;
    
    await searchPseudonymsWithQuery(pseudonymSearchQuery);
    updatePseudonymSearchURL(pseudonymSearchQuery, 1);
  };

  const handlePseudonymSearchClear = () => {
    setPseudonymSearchQuery("");
    setPseudonyms([]);
    setTotalPseudonyms(0);
    setPseudonymsTotalPages(1);
    updatePseudonymSearchURL("", 1);
    loadAllPseudonyms();
  };

  const updatePseudonymSearchURL = (query: string, page: number = 1) => {
    const newSearchParams = new URLSearchParams(searchParams);
    if (query.trim()) {
      newSearchParams.set('query', query);
    } else {
      newSearchParams.delete('query');
    }
    newSearchParams.set('pseudonymsPage', page.toString());
    router.replace(`/admin?tab=pseudonyms&${newSearchParams.toString()}`);
  };

  const handlePseudonymsPageChange = (newPage: number) => {
    setPseudonymsPage(newPage);
    updatePseudonymSearchURL(pseudonymSearchQuery, newPage);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  // Determine which pseudonyms to show and if we're in search mode
  const isSearchMode = pseudonymSearchQuery.trim() && pseudonyms.length > 0;
  const displayPseudonyms = isSearchMode ? pseudonyms : allPseudonyms;
  const isLoading = isSearchMode ? isPseudonymSearchLoading : isPseudonymsLoading;

  return (
    <div className="space-y-6">
      {/* Search Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Search className="h-5 w-5" />
            Search Pseudonyms
          </CardTitle>
          <CardDescription>
            Search for pseudonyms by display name, slug, or real identity
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handlePseudonymSearch} className="flex gap-2">
            <div className="space-y-4 w-full">
              <div>
                <Label htmlFor="pseudonym-search" className="text-sm font-medium">
                  Search for pseudonyms by display name, slug, or real identity
                </Label>
                <div className="mt-2 flex gap-2">
                  <Input
                    id="pseudonym-search"
                    placeholder="Search pseudonyms..."
                    value={pseudonymSearchQuery}
                    onChange={(e) => setPseudonymSearchQuery(e.target.value)}
                    className="flex-1"
                  />
                  <Button type="submit" disabled={isPseudonymSearchLoading || !pseudonymSearchQuery.trim()}>
                    {isPseudonymSearchLoading ? "Searching..." : "Search"}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handlePseudonymSearchClear}
                  >
                    Show All
                  </Button>
                </div>
              </div>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* Pseudonym List */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            Pseudonym Management
          </CardTitle>
          <CardDescription>
            {isSearchMode ? `Search results for "${pseudonymSearchQuery}"` : 'Manage platform pseudonyms and view details'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">
              <div className="text-muted-foreground">Loading pseudonyms...</div>
            </div>
          ) : (
            <>
              <div className="mb-4 flex items-center justify-between">
                <div className="text-sm text-muted-foreground">
                  {isSearchMode ? `Found ${pseudonyms.length} pseudonyms` : `Total Pseudonyms: ${totalPseudonyms}`}
                </div>
                <div className="text-sm text-muted-foreground">
                  Page {pseudonymsPage} of {pseudonymsTotalPages}
                </div>
              </div>
              
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Pseudonym ID</TableHead>
                    <TableHead>Display Name</TableHead>
                    <TableHead>Slug</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Karma</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {displayPseudonyms.map((pseudonym) => {
                    // Handle both SearchPseudonym and AdminPseudonymInfo types
                    const isSearchResult = 'pseudonymId' in pseudonym && typeof pseudonym.pseudonymId === 'string';
                    const isAdminResult = 'pseudonymId' in pseudonym && typeof pseudonym.pseudonymId === 'string';
                    
                    if (isSearchResult || isAdminResult) {
                      const adminPseudonym = pseudonym as AdminPseudonymInfo;
                      return (
                        <TableRow key={adminPseudonym.pseudonymId}>
                          <TableCell>
                            <Badge variant="outline" className="font-mono text-xs">
                              {adminPseudonym.pseudonymId.substring(0, 8)}...
                            </Badge>
                          </TableCell>
                          <TableCell>{adminPseudonym.displayName}</TableCell>
                          <TableCell>{adminPseudonym.slug}</TableCell>
                          <TableCell>
                            {adminPseudonym.isActive ? (
                              <Badge variant="secondary">Active</Badge>
                            ) : (
                              <Badge variant="outline">Inactive</Badge>
                            )}
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline">{adminPseudonym.karmaScore}</Badge>
                          </TableCell>
                          <TableCell>{formatDate(adminPseudonym.createdAt)}</TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Button 
                                variant="ghost" 
                                size="sm"
                                onClick={() => router.push(`/admin/pseudonyms/${adminPseudonym.pseudonymId}?returnTab=pseudonyms&returnQuery=${encodeURIComponent(pseudonymSearchQuery)}&returnPage=${pseudonymsPage}`)}
                                title="View pseudonym details"
                              >
                                <Eye className="h-4 w-4" />
                              </Button>
                              <Button variant="ghost" size="sm">
                                <Edit className="h-4 w-4" />
                              </Button>
                              <Button variant="ghost" size="sm">
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      );
                    }
                    
                    return null;
                  })}
                </TableBody>
              </Table>

              {/* Pagination */}
              {pseudonymsTotalPages > 1 && (
                <div className="flex items-center justify-center space-x-2 mt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePseudonymsPageChange(pseudonymsPage - 1)}
                    disabled={pseudonymsPage <= 1}
                  >
                    Previous
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    Page {pseudonymsPage} of {pseudonymsTotalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePseudonymsPageChange(pseudonymsPage + 1)}
                    disabled={pseudonymsPage >= pseudonymsTotalPages}
                  >
                    Next
                  </Button>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
