import { getApi } from './api-client';
import { AuthenticationApi, PseudonymsApi } from '@/generated/api/src/apis';

const authApi = getApi(AuthenticationApi);
const pseudonymsApi = getApi(PseudonymsApi);

export interface PseudonymOperationResult {
  success: boolean;
  data?: unknown;
  error?: string;
}

/**
 * Switch to a different pseudonym
 */
export async function switchPseudonym(pseudonymId: string): Promise<PseudonymOperationResult> {
  try {
    const response = await authApi.switchPseudonym({
      pseudonymId
    });
    
    return {
      success: true,
      data: response
    };
  } catch (error: unknown) {
    console.error('Failed to switch pseudonym:', error);
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to switch pseudonym'
    };
  }
}

/**
 * Create a new pseudonym
 */
export async function createPseudonym(displayName: string, bio?: string, websiteUrl?: string, slug?: string): Promise<PseudonymOperationResult> {
  try {
    const response = await pseudonymsApi.createPseudonym({
      displayName,
      bio: bio || '',
      websiteUrl: websiteUrl || '',
      allowDirectMessages: true,
      showKarma: true,
      slug: slug || ''
    });
    
    return {
      success: true,
      data: response
    };
  } catch (error: unknown) {
    console.error('Failed to create pseudonym:', error);
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to create pseudonym'
    };
  }
}

/**
 * Deactivate a pseudonym
 */
export async function deactivatePseudonym(pseudonymId: string): Promise<PseudonymOperationResult> {
  try {
    const response = await authApi.deactivatePseudonym({
      pseudonymId
    });
    
    return {
      success: true,
      data: response
    };
  } catch (error: unknown) {
    console.error('Failed to deactivate pseudonym:', error);
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to deactivate pseudonym'
    };
  }
}

/**
 * Get current user session with pseudonyms
 */
export async function getCurrentUserSession(): Promise<PseudonymOperationResult> {
  try {
    const response = await authApi.getCurrentUserSession();
    
    return {
      success: true,
      data: response
    };
  } catch (error: unknown) {
    console.error('Failed to get current user session:', error);
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to get user session'
    };
  }
} 