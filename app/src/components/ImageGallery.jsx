import React, { useState, useEffect } from 'react';
import { imageAPI, categoryAPI } from '../services/api';
import { formatDate } from '../utils/helpers';
import { jwtDecode } from 'jwt-decode';

const ImageGallery = () => {
  const [images, setImages] = useState([]);
  const [categories, setCategories] = useState([]);
  const [selectedCategory, setSelectedCategory] = useState('');
  const [newCategory, setNewCategory] = useState('');
  const [userRole, setUserRole] = useState('');
  const [loading, setLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [activeSearchTerm, setActiveSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(6);
  const [error, setError] = useState(''); // ✅ NEW

  // ✅ Decode token and extract user role
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      try {
        const decoded = jwtDecode(token);
        setUserRole(decoded?.role || '');
      } catch (err) {
        console.error('Error decoding token:', err);
      }
    }
  }, []);

  // ✅ Generic safe setter
  const safeSetImages = (fetchedImages, pages = 1) => {
    setImages(Array.isArray(fetchedImages) ? fetchedImages : []);
    setTotalPages(pages || 1);
  };

  // ✅ Fetch images
  const fetchImages = async (page = 1, limit = itemsPerPage) => {
    setLoading(true);
    setError('');
    try {
      const response = await imageAPI.getAll(page, limit);
      const { images: fetchedImages, totalPages: pages } = response.data || {};
      safeSetImages(fetchedImages, pages);
      setCurrentPage(page);
    } catch (error) {
      console.error('Error fetching images:', error);
      setError('Failed to load images.');
      setImages([]);
    } finally {
      setLoading(false);
    }
  };

  // ✅ Fetch images by category
  const fetchImagesByCategory = async (category, page = 1, limit = itemsPerPage) => {
    setLoading(true);
    setError('');
    try {
      const response = await imageAPI.getByCategory(category, page, limit);
      const { images: fetchedImages, totalPages: pages } = response.data || {};
      safeSetImages(fetchedImages, pages);
      setCurrentPage(page);
      if (!fetchedImages?.length) setError('No images found in this category.');
    } catch (error) {
      console.error('Error fetching by category:', error);
      setError('Error loading category images.');
      setImages([]);
    } finally {
      setLoading(false);
    }
  };

  // ✅ Search images
  const searchImages = async (term, page = 1, limit = itemsPerPage) => {
    setLoading(true);
    setError('');
    try {
      const response = await imageAPI.search(term, page, limit);
      const { images: fetchedImages, totalPages: pages } = response.data || {};
      safeSetImages(fetchedImages, pages);
      if (!fetchedImages?.length) setError('No images found for this search.');
      setCurrentPage(page);
    } catch (error) {
      console.error('Error searching images:', error);
      setError('Search failed. Try again.');
      setImages([]);
      setTotalPages(1);
    } finally {
      setLoading(false);
    }
  };

  // ✅ Fetch categories
  const fetchCategories = async () => {
    try {
      const response = await categoryAPI.getAll();
      setCategories(response.data?.Category || []);
    } catch (error) {
      console.error('Error fetching categories:', error);
    }
  };

  // ✅ Create category (admin only)
  const handleCreateCategory = async () => {
    if (!newCategory.trim()) return alert('Category name required');
    try {
      await categoryAPI.create({ category: newCategory });
      setNewCategory('');
      fetchCategories();
      alert('Category created successfully!');
    } catch (error) {
      console.error('Error creating category:', error);
      alert('Failed to create category');
    }
  };

  // ✅ Search handler
  const handleSearch = async () => {
    if (!searchTerm.trim()) {
      setActiveSearchTerm('');
      selectedCategory ? fetchImagesByCategory(selectedCategory, 1) : fetchImages(1);
      return;
    }
    setActiveSearchTerm(searchTerm);
    setSelectedCategory('');
    searchImages(searchTerm, 1);
  };

  // ✅ Pagination
  const handlePageChange = (newPage) => {
    if (activeSearchTerm) searchImages(activeSearchTerm, newPage);
    else if (selectedCategory) fetchImagesByCategory(selectedCategory, newPage);
    else fetchImages(newPage);
  };

  // ✅ Items per page
  const handleItemsPerPageChange = (newLimit) => {
    const limit = Number(newLimit);
    setItemsPerPage(limit);
    setCurrentPage(1);
    if (activeSearchTerm) searchImages(activeSearchTerm, 1, limit);
    else if (selectedCategory) fetchImagesByCategory(selectedCategory, 1, limit);
    else fetchImages(1, limit);
  };

  // ✅ Load data
  useEffect(() => {
    fetchImages();
  }, []);

  useEffect(() => {
    fetchCategories();
  }, [userRole]);

  return (
    <div className="image-gallery">
      <div className="gallery-controls">
        <div className="search-section">
          <input
            type="text"
            placeholder="Search images by name..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
          />
          <button onClick={handleSearch}>Search</button>
          {activeSearchTerm && (
            <button
              onClick={() => {
                setSearchTerm('');
                setActiveSearchTerm('');
                fetchImages(1);
              }}
              className="clear-search"
            >
              Clear
            </button>
          )}
        </div>

        <div className="filter-controls">
          {categories.length > 0 && (
            <div className="category-section">
              <label>Category:</label>
              <select
                value={selectedCategory}
                onChange={(e) => handleCategoryChange(e.target.value)}
                disabled={!!activeSearchTerm}
              >
                <option value="">All</option>
                {categories.map((cat, index) => (
                  <option key={index} value={cat.category}>
                    {cat.category}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="items-per-page-section">
            <label>Show:</label>
            <select
              value={itemsPerPage}
              onChange={(e) => handleItemsPerPageChange(e.target.value)}
            >
              <option value="6">6</option>
              <option value="12">12</option>
              <option value="24">24</option>
              <option value="48">48</option>
            </select>
          </div>
        </div>
      </div>

      {userRole === 'admin' && (
        <div className="create-category">
          <h3>Create Category</h3>
          <input
            type="text"
            placeholder="New category name"
            value={newCategory}
            onChange={(e) => setNewCategory(e.target.value)}
          />
          <button onClick={handleCreateCategory}>Add Category</button>
        </div>
      )}

      {loading ? (
        <div className="loading">Loading images...</div>
      ) : error ? (
        <div className="no-images">{error}</div>
      ) : images.length === 0 ? (
        <div className="no-images">No images found.</div>
      ) : (
        <>
          <div className="images-grid">
            {images.map((image) => (
              <div key={image.s3_key || image.id} className="image-card">
                <img
                  src={image.signed_url || image.s3_url}
                  alt={image.file_name}
                  loading="lazy"
                />
                <div className="image-info">
                  <h4>{image.file_name}</h4>
                  <p>Category: {image.category}</p>
                  <p>Uploaded: {formatDate(image.uploaded_at)}</p>
                </div>
              </div>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="pagination">
              <button
                disabled={currentPage === 1}
                onClick={() => handlePageChange(currentPage - 1)}
              >
                Previous
              </button>
              <span>
                Page {currentPage} of {totalPages}
              </span>
              <button
                disabled={currentPage === totalPages}
                onClick={() => handlePageChange(currentPage + 1)}
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default ImageGallery;
