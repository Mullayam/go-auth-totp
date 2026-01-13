import * as SecureStore from 'expo-secure-store';
import 'react-native-get-random-values';
import { v4 as uuidv4 } from 'uuid';

export interface Account {
    id: string;
    name: string;
    issuer: string;
    secret: string;
    type: 'totp';
}

const STORAGE_KEY = 'enjoys_authenticator_accounts';

export const saveAccount = async (account: Omit<Account, 'id'>) => {
    try {
        const existing = await getAccounts();
        // Simple deduplication check based on secret
        if (existing.some(a => a.secret === account.secret)) {
            return existing; // Or throw error
        }

        const newAccount: Account = {
            ...account,
            id: uuidv4(), // We might need to install uuid or just use crypto.randomUUID if available in RN context or a simple random string
        };

        // Polyfill for uuid if needed verify environment
        if (!newAccount.id) {
            newAccount.id = Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
        }

        const updated = [...existing, newAccount];
        await SecureStore.setItemAsync(STORAGE_KEY, JSON.stringify(updated));
        return updated;
    } catch (error) {
        console.error("Error saving account", error);
        throw error;
    }
};

export const getAccounts = async (): Promise<Account[]> => {
    try {
        const json = await SecureStore.getItemAsync(STORAGE_KEY);
        if (!json) return [];
        return JSON.parse(json);
    } catch (error) {
        console.error("Error getting accounts", error);
        return [];
    }
};

export const removeAccount = async (id: string) => {
    try {
        const existing = await getAccounts();
        const updated = existing.filter((a) => a.id !== id);
        await SecureStore.setItemAsync(STORAGE_KEY, JSON.stringify(updated));
        return updated;
    } catch (error) {
        console.error("Error removing account", error);
        throw error;
    }
};
