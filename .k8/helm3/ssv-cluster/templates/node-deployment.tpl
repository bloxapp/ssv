{{- define "node-deployment" }}
---
apiVersion: v1
kind: Service
metadata:
  name: node-{{ .Values.cluster_name }}-{{ .Values.node_name }}-svc
  namespace: {{ .Values.namespace }}
  labels:
    app: node-{{ .Values.cluster_name }}-{{ .Values.node_name }}
spec:
  type: ClusterIP
  ports:
    - port: 12001
      protocol: UDP
      targetPort: 12001
      name: port-12001
    - port: 13001
      protocol: TCP
      targetPort: 13001
      name: port-13001
    - port: 15001
      protocol: TCP
      targetPort: 15001
      name: port-15001
  selector:
    app: node-{{ .Values.cluster_name }}-{{ .Values.node_name }} 
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: node-{{ .Values.cluster_name }}-{{ .Values.node_name }}
  name: node-{{ .Values.cluster_name }}-{{ .Values.node_name }} 
  namespace: {{ .Values.namespace }}
spec:
  replicas: {{ .Values.replicaCount }}
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: node-{{ .Values.cluster_name }}-{{ .Values.node_name }}
  template:
    metadata:
      labels:
        app: node-{{ .Values.cluster_name }}-{{ .Values.node_name }}
    spec:
      containers:
      - name: node-{{ .Values.cluster_name }}-{{ .Values.node_name }} 
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        resources:
          limits:
            cpu:  {{ .Values.resources.limits.cpu }}
            memory: {{ .Values.resources.limits.mem }} 
        command: ["make", "start-node"]
        ports:
        - containerPort: 12001
          name: port-12001
          hostPort: 12001           
          protocol: UDP
        - containerPort: 13001
          name: port-13001
          hostPort: 13001
        - containerPort: 15001
          name: port-15001
          hostPort: 15001
        env:
        - name: CONFIG_PATH
          valueFrom:
            secretKeyRef:
              name: config-secrets
              key: config_path
        - name: BOOTNODES
          valueFrom:
            secretKeyRef:
              name: config-secrets
              key: boot_node
        - name: REGISTRY_CONTRACT_ADDR_KEY
          valueFrom:
            secretKeyRef:
              name: config-secrets
              key: smart_contract_addr_key
        - name: ETH_1_SYNC_OFFSET
          valueFrom:
            secretKeyRef:
              name: config-secrets
              key: eth_1_sync_offset
              optional: true
        - name: ABI_VERSION
          valueFrom:
            secretKeyRef:
              name: config-secrets
              key: abi_version
              optional: true
        - name: SHARE_CONFIG
          value: {{ .Values.envs.SHARE_CONFIG }}
        - name: LOG_LEVEL
          value: {{ .Values.envs.LOG_LEVEL }}
        - name: DEBUG_SERVICES
          value: {{ .Values.envs.DEBUG_SERVICES }}
        - name: DISCOVERY_TYPE_KEY
          value: {{ .Values.envs.DISCOVERY_TYPE_KEY }}
        - name: NETWORK
          value: {{ .Values.envs.NETWORK }}
        - name: CONSENSUS_TYPE
          value: {{ .Values.envs.CONSENSUS_TYPE }}
        - name: HOST_DNS
          value: {{ .Values.envs.HOST_DNS }}
        - name: HOST_ADDRESS
          value: {{ .Values.envs.HOST_ADDRESS }}
        - name: LOGGER_LEVEL
          value: {{ .Values.envs.LOGGER_LEVEL }}
        - name: DB_PATH
          value: {{ .Values.envs.DB_PATH }}
        - name: DB_REPORTING
          value: {{ .Values.envs.DB_REPORTING }}
        - name: ENABLE_PROFILE
          value: {{ .Values.envs.ENABLE_PROFILE }}
        - name: DISCOVERY_TRACE
          value: {{ .Values.envs.DISCOVERY_TRACE }}
        - name: PUBSUB_TRACE
          value: {{ .Values.envs.PUBSUB_TRACE }}
        - name: P2P_LOG
          value: {{ .Values.envs.P2P_LOG }}
        - name: NETWORK_ID
          value: {{ .Values.envs.NETWORK_ID }}
        - name: GENESIS_EPOCH
          value: {{ .Values.envs.GENESIS_EPOCH }}
        - name: CLEAN_REGISTRY_DATA
          value: {{ .Values.envs.CLEAN_REGISTRY_DATA }}
        volumeMounts:
        - mountPath: /data
          name: {{ .Values.VolumeName }}
        - mountPath: /data/share.yaml
          subPath: share.yaml
          name: ssv-cm-validator-options-ssv-node-cluster-{{ .Values.cluster_name }}-{{ .Values.node_name }}
      volumes:
      - name: {{ .Values.VolumeName }}
        persistentVolumeClaim:
          claimName: {{ .Values.VolumeName }}
      - name: ssv-cm-validator-options-ssv-node-cluster-{{ .Values.cluster_name }}-{{ .Values.node_name }} 
        configMap:
          name: ssv-cm-validator-options-ssv-node-cluster-{{ .Values.cluster_name }}-{{ .Values.node_name }} 
      hostNetwork: true
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}