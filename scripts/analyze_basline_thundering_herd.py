import json
from pathlib import Path
import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt

script_dir = Path(__file__).parent
results_dir = script_dir.parent / "results" / "formatted_locust_data"

# Load JSON files
with open(results_dir / 'baseline_target_scaling_test_results.json', 'r') as f:
    baseline = json.load(f)
    
with open(results_dir / 'thundering_herd_target_scaling_test_results.json', 'r') as f:
    thundering_herd = json.load(f)

# Test configuration data
test_configs = {
    'Baseline': {
        'workers': 1,
        'users': 50,
        'spawn_rate': 5,
        'run_time': '5min',
        'user_class': 'NormalUser',
        'wait_time': '1-3s',
        'scaling_policy': 'target_tracking'
    },
    'Thundering Herd': {
        'workers': 4,
        'users': 500,
        'spawn_rate': 500,
        'run_time': '3min',
        'user_class': 'AggressiveBuyer',
        'wait_time': '0.1-0.5s',
        'scaling_policy': 'target_tracking'
    }
}

# User class behaviors
user_behaviors = {
    'NormalUser': 'browse_product(35%), browse_catalog(30%), create_product(20%), health_check(10%), update_product(5%)',
    'AggressiveBuyer': 'update_product(50%), browse_product(30%), create_product(15%), browse_catalog(5%)'
}

# Extract summary data from raw results
def extract_summary(data, test_name):
    # Use raw results instead of summary
    df_results = pd.DataFrame(data['results'])
    
    rows = []
    for operation in df_results['operation'].unique():
        op_data = df_results[df_results['operation'] == operation]
        
        total_count = len(op_data)
        success_count = op_data['success'].sum()  # sum of boolean True values
        
        rows.append({
            'test': test_name,
            'operation': operation,
            'p50_response_time': op_data['response_time'].quantile(0.5),
            'avg_response_time': op_data['response_time'].mean(),
            'success_rate': (success_count / total_count * 100) if total_count > 0 else 0,
            'count': total_count,
            'failed_count': total_count - success_count
        })
    return pd.DataFrame(rows)

df_baseline = extract_summary(baseline, 'Baseline')
df_herd = extract_summary(thundering_herd, 'Thundering Herd')
df_combined = pd.concat([df_baseline, df_herd], ignore_index=True)

# Filter out health_check operations
df_combined = df_combined[df_combined['operation'] != 'health_check']

# Calculate error rate
df_combined['error_rate'] = 100 - df_combined['success_rate']

# ===== FIGURE 1: Test Configuration Table =====
fig1, ax_config = plt.subplots(figsize=(14, 3.5))
ax_config.axis('tight')
ax_config.axis('off')

config_data = []
for test, cfg in test_configs.items():
    user_class = cfg['user_class']
    # Split task distribution into two lines for better balance
    if user_class == 'NormalUser':
        task_dist = 'browse_product(35%), browse_catalog(30%),\ncreate_product(20%), health_check(10%), update_product(5%)'
    else:  # AggressiveBuyer
        task_dist = 'update_product(50%), browse_product(30%),\ncreate_product(15%), browse_catalog(5%)'
    
    config_data.append([
        test,
        cfg['users'],
        cfg['spawn_rate'],
        cfg['run_time'],
        user_class,
        cfg['wait_time'],
        task_dist,
        cfg['workers']
    ])

config_table = ax_config.table(
    cellText=config_data,
    colLabels=['Test', 'Users', 'Spawn\nRate', 'Duration', 'User Class', 'Wait\nTime', 'Task Distribution', 'Workers'],
    cellLoc='center',
    loc='center',
    colWidths=[0.15, 0.07, 0.07, 0.08, 0.15, 0.07, 0.45, 0.07]
)
config_table.auto_set_font_size(False)
config_table.set_fontsize(12)
config_table.scale(1, 4)

# Style the table
for i in range(len(config_data) + 1):
    for j in range(8):
        cell = config_table[(i, j)]
        if i == 0:
            cell.set_facecolor('#4a5568')
            cell.set_text_props(weight='bold', color='white')
        else:
            cell.set_facecolor('#f7fafc' if i % 2 == 0 else 'white')
            # Left-align task distribution column for readability
            if j == 6:
                cell.set_text_props(ha='left')
            
plt.title('Test Configuration', fontsize=14, fontweight='bold', pad=20)
plt.tight_layout()
plt.savefig('test_configuration.png', dpi=300, bbox_inches='tight')
plt.show()
print("Saved: test_configuration.png")

# ===== FIGURE 2: Performance Charts =====
sns.set_style("whitegrid")
fig2, axes = plt.subplots(2, 1, figsize=(12, 10))

# Plot 1: Response Time Comparison
bars1 = sns.barplot(data=df_combined, x='operation', y='p50_response_time', 
            hue='test', ax=axes[0], palette=["#2e40cc", "#e7923c"])
axes[0].set_title('P50 Response Time by Operation', fontsize=14, fontweight='bold')
axes[0].set_ylabel('P50 Response Time (ms)', fontsize=12)
axes[0].set_xlabel('Operation', fontsize=12)
axes[0].set_xticklabels(axes[0].get_xticklabels(), rotation=30, ha='right')
axes[0].legend(title='Test Type', loc='center left', bbox_to_anchor=(1, 0.5))

# Add value labels for P50
for i, bar in enumerate(axes[0].patches):
    height = bar.get_height()
    if height > 0:
        axes[0].text(bar.get_x() + bar.get_width()/2., height,
                    f'{height:.1f}ms',
                    ha='center', va='bottom', fontsize=9, fontweight='bold')

# Plot 2: Error Rate Comparison
operations_sorted = sorted(df_combined['operation'].unique())
tests_sorted = sorted(df_combined['test'].unique())

bars2 = sns.barplot(data=df_combined, x='operation', y='error_rate', 
            hue='test', ax=axes[1], palette=["#2e40cc", "#e7923c"])
axes[1].set_title('Error Rate by Operation', fontsize=14, fontweight='bold')
axes[1].set_ylabel('Error Rate (%)', fontsize=12)
axes[1].set_xlabel('Operation', fontsize=12)
axes[1].set_xticklabels(axes[1].get_xticklabels(), rotation=30, ha='right')
axes[1].legend(title='Test Type', loc='center left', bbox_to_anchor=(1, 0.5))

# Add detailed labels for error rates
# Get the bar containers from seaborn (one per test type)
for container_idx, container in enumerate(axes[1].containers):
    test_name = tests_sorted[container_idx] if container_idx < len(tests_sorted) else None
    
    for bar_idx, bar in enumerate(container):
        operation_name = operations_sorted[bar_idx] if bar_idx < len(operations_sorted) else None
        
        if test_name and operation_name:
            row = df_combined[(df_combined['operation'] == operation_name) & 
                             (df_combined['test'] == test_name)]
            
            if len(row) > 0:
                row = row.iloc[0]
                x_pos = bar.get_x() + bar.get_width()/2.
                
                if row['error_rate'] > 0:
                    label = f"{row['error_rate']:.2f}%\n({int(row['failed_count'])}/{int(row['count'])})"
                    y_pos = bar.get_height()
                else:
                    label = f"0%\n(0/{int(row['count'])})"
                    y_pos = axes[1].get_ylim()[1] * 0.02
                
                axes[1].text(x_pos, y_pos, label,
                            ha='center', va='bottom', fontsize=8, fontweight='bold')

plt.tight_layout()
plt.savefig('performance_comparison.png', dpi=300, bbox_inches='tight')
plt.show()
print("Saved: performance_comparison.png")

# ===== FIGURE 3: Response Time Over Time =====
fig3, axes_time = plt.subplots(2, 1, figsize=(14, 10))

# Prepare time series data from raw results
baseline_results = pd.DataFrame(baseline['results'])
thundering_results = pd.DataFrame(thundering_herd['results'])

# Convert timestamps to datetime
baseline_results['timestamp'] = pd.to_datetime(baseline_results['timestamp'])
thundering_results['timestamp'] = pd.to_datetime(thundering_results['timestamp'])

# Filter out health_check
baseline_results = baseline_results[baseline_results['operation'] != 'health_check']
thundering_results = thundering_results[thundering_results['operation'] != 'health_check']

# Calculate relative time in seconds from start
baseline_start = baseline_results['timestamp'].min()
thundering_start = thundering_results['timestamp'].min()

baseline_results['relative_time'] = (baseline_results['timestamp'] - baseline_start).dt.total_seconds()
thundering_results['relative_time'] = (thundering_results['timestamp'] - thundering_start).dt.total_seconds()

# Operation colors
colors = {'browse_catalog': '#e74c3c', 'browse_product': '#3498db', 
          'create_product': '#2ecc71', 'update_product': '#f39c12'}

# Plot 1: Baseline
operations = sorted(baseline_results['operation'].unique())
for operation in operations:
    op_data = baseline_results[baseline_results['operation'] == operation].sort_values('relative_time')
    
    if len(op_data) > 0:
        # Bin into 1-second intervals and calculate mean
        op_data['time_bin'] = (op_data['relative_time'] // 1).astype(int)
        binned = op_data.groupby('time_bin')['response_time'].mean().reset_index()
        
        axes_time[0].plot(binned['time_bin'], binned['response_time'], 
                    label=operation, 
                    color=colors.get(operation, '#95a5a6'),
                    linewidth=2, alpha=0.8)
        
        # Add markers for failed requests
        failed_data = op_data[op_data['success'] == False]
        if len(failed_data) > 0:
            failed_binned = failed_data.groupby('time_bin')['response_time'].mean().reset_index()
            axes_time[0].scatter(failed_binned['time_bin'], failed_binned['response_time'],
                               color=colors.get(operation, '#95a5a6'),
                               marker='x', s=100, linewidths=3, alpha=0.9, zorder=10)

axes_time[0].set_title('Baseline - Response Time Over Time (1s bins)', fontsize=14, fontweight='bold')
axes_time[0].set_ylabel('Response Time (ms)', fontsize=12)
axes_time[0].set_xlabel('Time (seconds)', fontsize=12)
axes_time[0].legend(loc='upper right', fontsize=10)
axes_time[0].grid(True, alpha=0.3)
axes_time[0].set_ylim(0, 400)

# Plot 2: Thundering Herd
operations = sorted(thundering_results['operation'].unique())
for operation in operations:
    op_data = thundering_results[thundering_results['operation'] == operation].sort_values('relative_time')
    
    if len(op_data) > 0:
        # Bin into 1-second intervals and calculate mean
        op_data['time_bin'] = (op_data['relative_time'] // 1).astype(int)
        binned = op_data.groupby('time_bin')['response_time'].mean().reset_index()
        
        axes_time[1].plot(binned['time_bin'], binned['response_time'], 
                    label=operation, 
                    color=colors.get(operation, '#95a5a6'),
                    linewidth=2, alpha=0.8)
        
        # Add markers for failed requests
        failed_data = op_data[op_data['success'] == False]
        if len(failed_data) > 0:
            failed_binned = failed_data.groupby('time_bin')['response_time'].mean().reset_index()
            axes_time[1].scatter(failed_binned['time_bin'], failed_binned['response_time'],
                               color=colors.get(operation, '#95a5a6'),
                               marker='x', s=100, linewidths=3, alpha=0.9, zorder=10)

axes_time[1].set_title('Thundering Herd - Response Time Over Time (1s bins)', fontsize=14, fontweight='bold')
axes_time[1].set_ylabel('Response Time (ms)', fontsize=12)
axes_time[1].set_xlabel('Time (seconds)', fontsize=12)
axes_time[1].legend(loc='upper right', fontsize=10)
axes_time[1].grid(True, alpha=0.3)
axes_time[1].set_ylim(0, 400)

plt.tight_layout()
plt.savefig('response_time_over_time.png', dpi=300, bbox_inches='tight')
plt.show()
print("Saved: response_time_over_time.png")

# Print summary tables
print("\n" + "="*80)
print("TEST CONFIGURATION SUMMARY")
print("="*80)
df_config = pd.DataFrame(test_configs).T
print(df_config.to_string())

print("\n" + "="*80)
print("USER CLASS BEHAVIORS")
print("="*80)
for user_class, behavior in user_behaviors.items():
    print(f"\n{user_class}:")
    print(f"  {behavior}")

print("\n" + "="*80)
print("P50 RESPONSE TIME SUMMARY (ms)")
print("="*80)
print(df_combined.pivot_table(
    values='p50_response_time', 
    index='operation', 
    columns='test'
).to_string())

print("\n" + "="*80)
print("AVERAGE RESPONSE TIME SUMMARY (ms)")
print("="*80)
print(df_combined.pivot_table(
    values='avg_response_time', 
    index='operation', 
    columns='test'
).to_string())

print("\n" + "="*80)
print("ERROR RATE SUMMARY (%)")
print("="*80)
print(df_combined.pivot_table(
    values='error_rate', 
    index='operation', 
    columns='test'
).to_string())

print("\n" + "="*80)
print("REQUEST COUNT BY OPERATION")
print("="*80)
print(df_combined.pivot_table(
    values='count', 
    index='operation', 
    columns='test',
    aggfunc='sum'
).to_string())
print("\n")