import pandas as pd
import pickle
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler, OneHotEncoder
from sklearn.compose import ColumnTransformer
from sklearn.metrics import (
    roc_auc_score, precision_score, recall_score, f1_score, accuracy_score,
    precision_recall_curve, roc_curve, auc
)
from catboost import CatBoostClassifier
from sklearn.svm import SVC
from sklearn.linear_model import LogisticRegression
from sklearn.neighbors import KNeighborsClassifier
from sklearn.ensemble import RandomForestClassifier
from xgboost import XGBClassifier
from lightgbm import LGBMClassifier
from sklearn.neural_network import MLPClassifier
import numpy as np
import matplotlib.pyplot as plt

# 读取数据
data = pd.read_csv(r'D:\PyCharmproject\code\kidney\dixueya\透析数据_拉平+基本信息.csv')
data['LogTime_2'] = pd.to_datetime(data['LogTime_2'])

# 创建目标变量
data['LowBloodPressure'] = np.where(
    (data['lable_SBP_change'] == 1) |
    (data['lable_MAP_change'] == 1) |
    (data['lable_SBP_low'] == 1), 1, 0
)

# 删除 2024.8.12 前正常血压的数据
data = data[~((data['LogTime_2'] <= '2024-08-12') & (data['LowBloodPressure'] == 0))]

# 处理缺失值
numeric_cols = data.select_dtypes(include=[np.number]).columns
data[numeric_cols] = data[numeric_cols].fillna(data[numeric_cols].mean())

# 删除多余列
drop_cols = [f'LogTime_{i}' for i in range(2, 31)] + ['TimeDiff']
data = data.drop(columns=drop_cols)

print("数据当前包含的列：")
for col in data.columns:
    print(col)

# 提取特征和标签
features = data.drop(['MeasurementSessionId', 'TreatmentId', 'PatientId','TreatmentId_mean','TreatmentId_std',
                      'lable_SBP_change', 'lable_MAP_change', 'lable_SBP_low',
                      'lable_SBP_high', 'LowBloodPressure', 'TMP_StdDev', 'lable_TMP',
                      'lable_DBP_high',], axis=1)
target = data['LowBloodPressure']

# --- 处理分类和数值特征分开编码 ---
categorical_cols = ['DialysisMethod']
numerical_cols = features.columns.difference(categorical_cols)

# 构建列转换器
preprocessor = ColumnTransformer(
    transformers=[
        ('num', StandardScaler(), numerical_cols),
        ('cat', OneHotEncoder(drop='first'), categorical_cols)  # drop='first' 避免多重共线性
    ]
)

# 拟合并转换
features_scaled_array = preprocessor.fit_transform(features)

# 构建新特征列名
ohe = preprocessor.named_transformers_['cat']
cat_col_names = ohe.get_feature_names_out(categorical_cols)
all_columns = list(numerical_cols) + list(cat_col_names)

# 构建标准化后的 DataFrame
features_scaled = pd.DataFrame(features_scaled_array.toarray()
                               if hasattr(features_scaled_array, 'toarray')
                               else features_scaled_array,
                               columns=all_columns)
# 划分训练测试集
X_train, X_test, y_train, y_test = train_test_split(
    features_scaled, target, test_size=0.2, random_state=42, stratify=target
)

# 模型定义
models = {
    "CatBoost": CatBoostClassifier(
        iterations=1000, learning_rate=0.03, depth=12,
        task_type='CPU', random_state=42, verbose=0
    ),
    "SVM": SVC(kernel='rbf', C=2.0, gamma='scale', probability=True, random_state=42),
    "Logistic Regression": LogisticRegression(solver='liblinear', C=1.0, random_state=42),
    "KNN": KNeighborsClassifier(n_neighbors=7),
    "Random Forest": RandomForestClassifier(
        n_estimators=300, max_depth=15, random_state=42, class_weight='balanced'
    ),
    "XGBoost": XGBClassifier(
        n_estimators=300, max_depth=10, learning_rate=0.05,
        subsample=0.8, colsample_bytree=0.8, random_state=42
    ),
    "LightGBM": LGBMClassifier(
        n_estimators=300, num_leaves=50, learning_rate=0.05, random_state=42
    ),
    "MLP": MLPClassifier(hidden_layer_sizes=(100, 50), max_iter=300, random_state=42)
}

# 训练并保存模型
results = []
roc_curves = {}
prc_curves = {}

for name, model in models.items():
    print(f"Training {name}...")
    model.fit(X_train, y_train)
    y_prob = model.predict_proba(X_test)[:, 1]
    y_pred_bin = (y_prob > 0.5).astype(int)

    acc = accuracy_score(y_test, y_pred_bin)
    prec = precision_score(y_test, y_pred_bin, zero_division=0)
    rec = recall_score(y_test, y_pred_bin)
    f1 = f1_score(y_test, y_pred_bin)
    roc_auc = roc_auc_score(y_test, y_prob)
    precision_arr, recall_arr, _ = precision_recall_curve(y_test, y_prob)
    prc_auc = auc(recall_arr, precision_arr)

    results.append({
        'Model': name,
        'Accuracy': acc,
        'Precision': prec,
        'Recall': rec,
        'F1 Score': f1,
        'AUC': roc_auc,
        'PRC AUC': prc_auc
    })

    fpr, tpr, _ = roc_curve(y_test, y_prob)
    roc_curves[name] = (fpr, tpr, roc_auc)
    prc_curves[name] = (recall_arr, precision_arr, prc_auc)

    with open(f"{name.replace(' ', '_')}_model.pkl", "wb") as f:
        pickle.dump(model, f)

# 保存缩放器
with open("scaler.pkl", "wb") as f:
    pickle.dump(preprocessor, f)

# 打印评估结果
results_df = pd.DataFrame(results).sort_values(by='AUC', ascending=False)
print("\n模型性能指标：")
print(results_df.to_string(index=False, float_format="%.4f"))

# 绘制 ROC 曲线
plt.figure(figsize=(10, 8))
for name, (fpr, tpr, roc_auc_val) in roc_curves.items():
    plt.plot(fpr, tpr, label=f'{name} (AUC={roc_auc_val:.3f})')
plt.plot([0, 1], [0, 1], 'k--', lw=1)
plt.xlabel('False Positive Rate')
plt.ylabel('True Positive Rate')
plt.title('ROC Curve Comparison')
plt.legend(loc='lower right')
plt.grid(True)
plt.tight_layout()
plt.show()

# 绘制 PRC 曲线
plt.figure(figsize=(10, 8))
for name, (recall, precision, prc_auc_val) in prc_curves.items():
    plt.plot(recall, precision, label=f'{name} (AUC={prc_auc_val:.3f})')
plt.xlabel('Recall')
plt.ylabel('Precision')
plt.title('Precision-Recall Curve Comparison')
plt.legend(loc='lower left')
plt.grid(True)
plt.tight_layout()
plt.show()
