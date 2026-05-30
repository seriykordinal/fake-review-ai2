
class TestHealth:
    def test_health_loaded(self, client):
        c, _ = client
        resp = c.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "ok"
        assert data["model_loaded"] is True

    def test_health_not_loaded(self, client_model_not_loaded):
        c, _ = client_model_not_loaded
        resp = c.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "loading"
        assert data["model_loaded"] is False



class TestPredict:
    def test_predict_success(self, client):
        c, mock_pred = client
        resp = c.post("/predict", json={"text": "Отличный товар"})
        assert resp.status_code == 200
        data = resp.json()
        assert data["fake_probability"] == 0.75
        mock_pred.predict_one.assert_called_once_with(text="Отличный товар", rating=None)

    def test_predict_with_rating(self, client):
        c, mock_pred = client
        resp = c.post("/predict", json={"text": "Текст", "rating": "5"})
        assert resp.status_code == 200
        mock_pred.predict_one.assert_called_once_with(text="Текст", rating="5")

    def test_predict_empty_text_rejected(self, client):
        c, _ = client
        resp = c.post("/predict", json={"text": ""})
        assert resp.status_code == 422  

    def test_predict_missing_text(self, client):
        c, _ = client
        resp = c.post("/predict", json={})
        assert resp.status_code == 422

    def test_predict_invalid_json(self, client):
        c, _ = client
        resp = c.post("/predict", content="not json",
                       headers={"Content-Type": "application/json"})
        assert resp.status_code == 422

    def test_predict_model_not_loaded(self, client_model_not_loaded):
        c, _ = client_model_not_loaded
        resp = c.post("/predict", json={"text": "Текст"})
        assert resp.status_code == 503



class TestPredictBatch:
    def test_batch_success(self, client):
        c, mock_pred = client
        texts = ["Отзыв 1", "Отзыв 2", "Отзыв 3"]
        resp = c.post("/predict_batch", json={"texts": texts})
        assert resp.status_code == 200
        data = resp.json()
        assert data["probabilities"] == [0.5, 0.6, 0.7]
        mock_pred.predict_batch.assert_called_once_with(texts=texts)

    def test_batch_empty_list(self, client):
        c, _ = client
        resp = c.post("/predict_batch", json={"texts": []})
        assert resp.status_code == 200
        data = resp.json()
        assert data["probabilities"] == []

    def test_batch_missing_texts(self, client):
        c, _ = client
        resp = c.post("/predict_batch", json={})
        assert resp.status_code == 422

    def test_batch_model_not_loaded(self, client_model_not_loaded):
        c, _ = client_model_not_loaded
        resp = c.post("/predict_batch", json={"texts": ["Текст"]})
        assert resp.status_code == 503
