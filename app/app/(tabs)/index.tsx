import { CodeCard } from '@/components/CodeCard';
import { Footer } from '@/components/Footer';
import { PremiumTheme } from '@/constants/PremiumTheme';
import { Account, getAccounts } from '@/utils/storage';
import { MaterialIcons } from '@expo/vector-icons';
import { LinearGradient } from 'expo-linear-gradient';
import { useFocusEffect, useRouter } from 'expo-router';
import { useCallback, useState } from 'react';
import { FlatList, StatusBar, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

export default function HomeScreen() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const router = useRouter();

  const loadAccounts = useCallback(async () => {
    const data = await getAccounts();
    setAccounts(data);
  }, []);

  useFocusEffect(
    useCallback(() => {
      loadAccounts();
    }, [loadAccounts])
  );

  return (
    <LinearGradient
      colors={[PremiumTheme.colors.backgroundGradientStart, PremiumTheme.colors.backgroundGradientEnd]}
      style={styles.container}
    >
      <StatusBar barStyle="light-content" />
      <SafeAreaView style={styles.safeArea}>
        <View style={styles.header}>
          <View>
            <Text style={styles.subtitle}>Welcome back</Text>
            <Text style={styles.title}>Authenticator</Text>
          </View>
          <View style={styles.headerIcon}>
            <MaterialIcons name="security" size={24} color={PremiumTheme.colors.primary} />
          </View>
        </View>

        <FlatList
          data={accounts}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => <CodeCard account={item} />}
          contentContainerStyle={styles.listContent}
          showsVerticalScrollIndicator={false}
          ListEmptyComponent={
            <View style={styles.emptyState}>
              <MaterialIcons name="lock-outline" size={64} color={PremiumTheme.colors.textSecondary} style={{ opacity: 0.5 }} />
              <Text style={styles.emptyText}>No accounts yet</Text>
              <Text style={styles.emptySubText}>Scan a QR code to get started</Text>
            </View>
          }
        />

        <TouchableOpacity
          style={styles.fab}
          onPress={() => router.push('/scan')}
          activeOpacity={0.8}
        >
          <LinearGradient
            colors={[PremiumTheme.colors.primaryGradientStart, PremiumTheme.colors.primaryGradientEnd]}
            style={styles.fabGradient}
          >
            <MaterialIcons name="add" size={32} color="white" />
          </LinearGradient>
        </TouchableOpacity>

        <Footer />
      </SafeAreaView>
    </LinearGradient>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  safeArea: {
    flex: 1,
  },
  header: {
    paddingHorizontal: 24,
    paddingTop: 20,
    paddingBottom: 20,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  subtitle: {
    fontSize: 14,
    color: PremiumTheme.colors.primary,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 1,
    marginBottom: 4,
  },
  title: {
    fontSize: 32,
    fontWeight: '800',
    color: PremiumTheme.colors.text,
    letterSpacing: -1,
  },
  headerIcon: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: 'rgba(255,255,255,0.05)',
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  listContent: {
    paddingBottom: 100,
    paddingTop: 10,
  },
  emptyState: {
    padding: 60,
    alignItems: 'center',
    justifyContent: 'center',
  },
  emptyText: {
    fontSize: 20,
    fontWeight: '700',
    color: PremiumTheme.colors.text,
    marginTop: 16,
    marginBottom: 8,
  },
  emptySubText: {
    fontSize: 14,
    color: PremiumTheme.colors.textSecondary,
  },
  fab: {
    position: 'absolute',
    bottom: 30,
    right: 30,
    ...PremiumTheme.shadows.glow,
  },
  fabGradient: {
    width: 64,
    height: 64,
    borderRadius: 32,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.2)',
  }
});
